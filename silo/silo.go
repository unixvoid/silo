package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/unixvoid/glogger"
	"gopkg.in/gcfg.v1"
	"gopkg.in/redis.v5"
)

type Config struct {
	Silo struct {
		Loglevel       string
		Port           int
		Content        string
		Domain         string
		BaseDir        string
		BootstrapDelay time.Duration
		PollDelay      time.Duration
	}
	SSL struct {
		UseTLS     bool
		ServerCert string
		ServerKey  string
	}
	Redis struct {
		Host        string
		Password    string
		CleanOnBoot bool
	}
}

var (
	config = Config{}
)

func main() {
	// read in config file
	readConf()

	// initialize the logger with the configured loglevel
	initLogger(config.Silo.Loglevel)

	// initialize redis connection
	redisClient, err := initRedisConnection()
	if err != nil {
		glogger.Debug.Printf("redis conneciton cannot be made, trying again in %d seconds", config.Silo.BootstrapDelay)
		time.Sleep(config.Silo.BootstrapDelay * time.Second)
		redisClient, err = initRedisConnection()
		if err != nil {
			glogger.Error.Println("redis connection cannot be made.")
			os.Exit(1)
		}
	}
	glogger.Debug.Println("connection to redis succeeded.")
	glogger.Info.Println("link to redis on", config.Redis.Host)

	// clean redis on boot if its set in the config
	if config.Redis.CleanOnBoot {
		glogger.Debug.Println("cleaning redis")
		redisClient.FlushAll()
	}

	// populate redis with available packages
	go populatePackages(config.Silo.PollDelay, config.Silo.Content, config.Silo.Domain, config.Silo.BaseDir, redisClient)

	// handle web requests/routes
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		glogger.Debug.Printf("routing: %s\n", r.RequestURI)
		serveroot(w, r, redisClient)
	}).Methods("GET")
	router.HandleFunc("/rkt/{project}/{pkg}", func(w http.ResponseWriter, r *http.Request) {
		glogger.Debug.Printf("routing: %s\n", r.RequestURI)
		handlerdynamic(w, r, redisClient)
	}).Methods("GET")
	router.HandleFunc("/{*}", func(w http.ResponseWriter, r *http.Request) {
		glogger.Debug.Printf("routing: %s\n", r.RequestURI)
		handlerwildcard(w, r, redisClient)
	}).Methods("GET")

	if config.SSL.UseTLS {
		tlsConfig := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
			ClientSessionCache: tls.NewLRUClientSessionCache(128),
		}
		glogger.Info.Println("silo running https on", config.Silo.Port)
		tlsServer := &http.Server{Addr: fmt.Sprintf(":%d", config.Silo.Port), Handler: router, TLSConfig: tlsConfig}
		log.Fatal(tlsServer.ListenAndServeTLS(config.SSL.ServerCert, config.SSL.ServerKey))
	} else {
		glogger.Info.Println("silo running http on", config.Silo.Port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Silo.Port), router))
	}
}

func readConf() {
	// init config file
	err := gcfg.ReadFileInto(&config, "config.gcfg")
	if err != nil {
		panic(fmt.Sprintf("Could not load config.gcfg, error: %s\n", err))
	}
}

func initLogger(logLevel string) {
	// init logger
	if logLevel == "debug" {
		glogger.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else if logLevel == "cluster" {
		glogger.LogInit(os.Stdout, os.Stdout, ioutil.Discard, os.Stderr)
	} else if logLevel == "info" {
		glogger.LogInit(os.Stdout, ioutil.Discard, ioutil.Discard, os.Stderr)
	} else {
		glogger.LogInit(ioutil.Discard, ioutil.Discard, ioutil.Discard, os.Stderr)
	}
}

func initRedisConnection() (*redis.Client, error) {
	// init redis connection
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Host,
		Password: config.Redis.Password,
		DB:       0,
	})

	_, redisErr := redisClient.Ping().Result()
	return redisClient, redisErr
}

func serveroot(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	// get metadata from redis
	metadata, err := redisClient.Get("master:metadata").Result()
	if err != nil {
		glogger.Debug.Println("error getting master:metadata for root display")
		glogger.Debug.Println(err)
	}

	// generate the html page
	page := fmt.Sprintf("<html>\n<head>%s\n</head>\n</html>", metadata)
	// serve metadata to client
	fmt.Fprintf(w, page)
}

func handlerdynamic(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	vars := mux.Vars(r)
	project := vars["project"]
	pkg := vars["pkg"]

	// see if the artifact exists
	exists, err := redisClient.SIsMember("master:packages", project).Result()
	if err != nil {
		glogger.Error.Println("error getting result from master:packages")
		glogger.Error.Println(err)
	}

	if exists {
		// serve up file
		servePackage(config.Silo.BaseDir, project, pkg, w, r)
		//fmt.Fprintf(w, "file here :D %s:%s", project, pkg)
	} else {
		glogger.Debug.Printf("data '%s' does not exist\n", project)
		w.WriteHeader(http.StatusNotFound)
		// TODO display file not found message
	}
}

func handlerwildcard(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	// TODO make a default response of 404 if no other criteria exist
	//   mux already handles it with a 404.. but maybe a nice message?

	if strings.Contains(r.RequestURI, "ac-discovery") {
		serveroot(w, r, redisClient)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
