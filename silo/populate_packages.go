package main

import (
	"fmt"
	"io/ioutil"

	"github.com/unixvoid/glogger"
	"gopkg.in/redis.v5"
)

func populatePackages(domain, basedir string, redisClient *redis.Client) {
	dirs, _ := ioutil.ReadDir(basedir)
	for _, d := range dirs {
		// dont add 'pubkey', we are not indexing it
		if d.Name() != "pubkey" {
			glogger.Debug.Printf("adding entry to 'master:packages' :: %s", d.Name())
			err := redisClient.SAdd("master:packages", d.Name()).Err()
			if err != nil {
				glogger.Error.Printf("error adding entry '%s' to master:packages\n", d.Name())
				glogger.Error.Println(err)
			} else {
				// if entry was added to master:packages, generate the metadata line in another thread
				go generateMeta(domain, d.Name(), redisClient)
			}
		}
	}
}

func generateMeta(domain, pkg string, redisClient *redis.Client) {
	//aciline := fmt.Sprintf("<meta name=\"ac-discovery\" content=\"%s/%s https://%s/rkt/%s/%s-{version}-{os}-{arch}.{ext}\">", domain, pkg, domain, pkg, pkg)
	//pubkeyline := fmt.Sprintf("<meta name=\"ac-discovery-pubkeys\" content=\"%s/%s https://%s/rkt/pubkey/pubkeys.gpg\">", domain, pkg, domain)
	//metaentry := fmt.Sprintf("%s\n%s", aciline, pubkeyline)

	metaentry := fmt.Sprintf(`<meta name="ac-discovery" content="%s/%s https://%s/rkt/%s/%s-{version}-{os}-{arch}.{ext}">
<meta name="ac-discovery-pubkeys" content="%s/%s https://%s/rkt/pubkey/pubkeys.gpg">`, domain, pkg, domain, pkg, pkg, domain, pkg, domain)

	glogger.Debug.Printf("adding meta line to 'package:%s'", pkg)
	err := redisClient.Set(fmt.Sprintf("package:%s", pkg), metaentry, 0).Err()
	if err != nil {
		glogger.Error.Printf("error adding meta tag to package:%s\n", pkg)
		glogger.Error.Println(err)
	}
}
