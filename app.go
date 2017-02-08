package main

import (
	"archive/tar"
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var port string
var dockercli, err = client.NewClient("unix:///var/run/docker.sock", "v1.25", nil, map[string]string{"User-Agent": "nanoserverless"})
var tagprefix = "nanoserverless"

func init() {
	flag.StringVar(&port, "port", "80", "give me a port number")
}

func main() {
	// defer profile.Start().Stop()
	flag.Parse()

	// Docker

	// Router
	r := mux.NewRouter()
	r.HandleFunc("/whoami", whoami)
	r.HandleFunc("/create/{base}/{name}", create)
	http.Handle("/", r)
	fmt.Println("Starting up on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func whoami(w http.ResponseWriter, req *http.Request) {
	u, _ := url.Parse(req.URL.String())
	queryParams := u.Query()
	wait := queryParams.Get("wait")
	if len(wait) > 0 {
		duration, err := time.ParseDuration(wait)
		if err == nil {
			time.Sleep(duration)
		}
	}

	hostname, _ := os.Hostname()
	fmt.Fprintln(w, "Hostname:", hostname)

	req.Write(w)
}

func exec(w http.ResponseWriter, req *http.Request) {
	/*vars := mux.Vars(req)
	base := vars["base"]
	name := vars["name"]
	tag := tagprefix + "-" + base + "-" + name*/
	fmt.Fprintln(w, "Not implemented yet")
}

func create(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	base := vars["base"]
	name := vars["name"]
	tag := tagprefix + "-" + base + "-" + name
	bodyb, _ := ioutil.ReadAll(req.Body)
	body := string(bodyb)

	// Generate dockerfile
	dockerfile := "FROM "
	if base == "php7" {
		dockerfile += "php:7"
	}
	if base == "node7" {
		dockerfile += "node:7"
	}
	dockerfile += "\nCOPY app /"
	if base == "php7" {
		dockerfile += "\nENTRYPOINT [\"php\", \"app\"]"
	}
	if base == "node7" {
		dockerfile += "\nENTRYPOINT [\"node\", \"app\"]"
	}

	// Generate app
	app := ""
	if base == "php7" {
		//app += "<?php\nparse_str($argv[1], $params);\n"
		app += "<?php\n"
	}
	/*if base == "node7" {
		app += "var querystring = require('querystring');\nvar params = querystring.parse(process.argv[2]);\n"
	}*/
	app += body + "\n"
	if base == "php7" {
		app += "?>"
	}

	// Buffer context
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	// Add some files
	var files = []struct {
		Name, Body string
	}{
		{"Dockerfile", dockerfile},
		{"app", app},
	}
	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			log.Fatalln(err)
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			log.Fatalln(err)
		}
	}

	// Make sure to check the error on Close.
	if err := tw.Close(); err != nil {
		log.Fatalln(err)
	}

	// Open the tar archive for reading.
	reader := bytes.NewReader(buf.Bytes())
	// Docker build
	buildOptions := types.ImageBuildOptions{
		Tags:           []string{tag},
		NoCache:        true,
		SuppressOutput: true,
	}

	response, err := dockercli.ImageBuild(context.Background(), reader, buildOptions)
	if err != nil {
		fmt.Fprintln(w, "Error in creating image", tag)
	}
	defer response.Body.Close()
	//fmt.Fprintf(dockercli.Out(), "%s", response.Body)
	buf2 := new(bytes.Buffer)
	buf2.ReadFrom(response.Body)
	result := buf2.String()
	fmt.Fprintln(w, "Image ", tag, "created !\n")
	fmt.Fprintln(w, "Dockerfile:")
	fmt.Fprintln(w, dockerfile, "\n")
	fmt.Fprintln(w, "Code:")
	fmt.Fprintln(w, app, "\n")
	fmt.Fprintln(w, "Log:")
	fmt.Fprintln(w, result, "\n")
	//fmt.Fprintln(w, "response:", response.Body)
	//buildCtx := ioutil.NopCloser(reader)
	//dockercli.ImageBuild(context.Background(), buildCtx, buildOptions)
}
