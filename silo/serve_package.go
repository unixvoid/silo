package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/unixvoid/glogger"
)

func servePackage(basedir, project, pkg string, w http.ResponseWriter, r *http.Request) {
	// test project dir
	//   if it does not exist pop 500
	//   this should exist.. we tested with redis before
	projectPath := fmt.Sprintf("%s/%s", basedir, project)
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		// project does not exist, 500 and exit
		glogger.Debug.Printf("project '%s' does not exist\n", project)
		glogger.Debug.Printf("tested '%s'\n", projectPath)

		// return 500 to client, project should exist but now doesn't
		//   did the filesystem change since start?
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// test package dir
	//   if it does not exist pop 404
	packagePath := fmt.Sprintf("%s/%s/%s", basedir, project, pkg)
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		// package does not exist, 404 and exit
		glogger.Debug.Printf("package '%s' does not exist\n", pkg)
		glogger.Debug.Printf("tested '%s'\n", packagePath)

		// return 404 to the client, the specified pacakge does not exist
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// the package exists, serve it
	glogger.Debug.Printf("serving '%s'\n", packagePath)
	http.ServeFile(w, r, packagePath)
}
