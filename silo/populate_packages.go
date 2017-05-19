package main

import (
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/unixvoid/glogger"
	"gopkg.in/redis.v5"
)

func populatePackages(uselivereload bool, polldelay time.Duration, content, domain, basedir string, redisClient *redis.Client) {
	dirs, _ := ioutil.ReadDir(basedir)
	// make a wait group for concurrency
	var wg sync.WaitGroup

	for _, d := range dirs {
		// dont add 'pubkey', we are not indexing it
		if d.Name() != "pubkey" {
			// up the waitgroup counter
			wg.Add(1)
			glogger.Debug.Printf("adding entry to 'master:packages' :: %s", d.Name())
			err := redisClient.SAdd("master:packages", d.Name()).Err()
			if err != nil {
				glogger.Error.Printf("error adding entry '%s' to master:packages\n", d.Name())
				glogger.Error.Println(err)
			} else {
				// if entry was added to master:packages, generate the metadata line
				go generateMeta(&wg, content, domain, d.Name(), redisClient)
			}
		}
	}
	// wait for all concurrent processes to complete
	wg.Wait()
	// now that the data has been entered, generate the master metadata header
	glogger.Debug.Println("updating master:metadata")
	packages, err := redisClient.SInter("master:packages").Result()
	if err != nil {
		glogger.Error.Println("error getting master:package for metadata concatination")
		glogger.Error.Println(err)
	}
	for _, pkg := range packages {
		// get current master metadata
		currentMeta, _ := redisClient.Get("master:metadata").Result()
		// get package metadata
		pkgMeta, _ := redisClient.Get(fmt.Sprintf("package:%s", pkg)).Result()
		newMeta := fmt.Sprintf("%s\n%s", currentMeta, pkgMeta)

		err := redisClient.Set("master:metadata", newMeta, 0).Err()
		if err != nil {
			glogger.Error.Printf("error updating master:meatadata with '%s'\n", pkg)
			glogger.Error.Println(err)
		}
	}

	// we are done populating master:packages, run the filesystem watcher
	// TODO eventually this function will die and it will only be the fswatcher diff
	//   the diff will run first try and see that master:packages is empty and populate it
	if uselivereload {
		for {
			go fsWatcher(basedir, redisClient)
			time.Sleep(polldelay * time.Second)
		}
	}
}

func generateMeta(wg *sync.WaitGroup, content, domain, pkg string, redisClient *redis.Client) {
	metaentry := fmt.Sprintf(``+
		`<meta name="ac-discovery" content="%s/%s %s/rkt/%s/%s-{version}-{os}-{arch}.{ext}">`+"\n"+
		`<meta name="ac-discovery-pubkeys" content="%s/%s %s/rkt/pubkey/pubkeys.gpg">`,
		content, pkg, domain, pkg, pkg, content, pkg, domain)

	glogger.Debug.Printf("adding meta line to 'package:%s'", pkg)
	err := redisClient.Set(fmt.Sprintf("package:%s", pkg), metaentry, 0).Err()
	if err != nil {
		glogger.Error.Printf("error adding meta tag to package:%s\n", pkg)
		glogger.Error.Println(err)
	}

	// decrement the wg counter
	defer wg.Done()
}

func fsWatcher(basedir string, redisClient *redis.Client) {
	// populate live:packages with fs footprint
	//glogger.Debug.Printf("footprinting fs at '%s'\n", basedir)
	dirs, _ := ioutil.ReadDir(basedir)

	for _, d := range dirs {
		// dont add 'pubkey', we are not indexing it
		if d.Name() != "pubkey" {
			//glogger.Debug.Printf("adding entry to 'live:packages' :: %s", d.Name())
			err := redisClient.SAdd("live:packages", d.Name()).Err()
			if err != nil {
				glogger.Error.Printf("error adding entry '%s' to live:packages\n", d.Name())
				glogger.Error.Println(err)
			}
		}
	}
	// call diff function to diff live:packages against master:packages
	go packageDiff(redisClient)
}

func packageDiff(redisClient *redis.Client) {
	// diff master:packages against live:packages

	// diff for added packages
	diffString, _ := redisClient.SDiff("live:packages", "master:packages").Result()
	for _, b := range diffString {
		glogger.Debug.Printf("adding entry to 'master:packages' :: %s", b)
		err := redisClient.SAdd("master:packages", b).Err()
		if err != nil {
			glogger.Error.Printf("error adding entry '%s' to master:packages\n", b)
			glogger.Error.Println(err)
		} else {
			// rebuild metadata
		}
	}

	// diff for removed packages
	diffString, _ = redisClient.SDiff("master:packages", "live:packages").Result()
	for _, b := range diffString {
		glogger.Debug.Printf("removing entry from 'master:packages' :: %s", b)
		err := redisClient.SRem("master:packages", b).Err()
		if err != nil {
			glogger.Error.Printf("error removing entry '%s' from master:packages\n", b)
			glogger.Error.Println(err)
		} else {
			// rebuild metadata
		}
	}

	// done diffing, clear out live set
	redisClient.Del("live:packages")
}
