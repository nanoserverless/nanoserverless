package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
	"context"
	"github.com/gorilla/mux"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"archive/tar"
	"bytes"
	"io"
)

var port string
var dockercli, err = client.NewClient("unix:///var/run/docker.sock", "v1.25", nil, map[string]string{"User-Agent": "nanoserverless"})

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

func create(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	base := vars["base"]
	name := vars["name"]
	fmt.Fprintln(w, "base:", base)
	fmt.Fprintln(w, "name:", name)
	fmt.Fprintln(w, "docker cli:", dockercli.ClientVersion())

	containers, err := dockercli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Fprintln(w, container.ID[:10], container.Image)
	}

	// Buffer context
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	// Add some files
	var files = []struct {
		Name, Body string
	}{
		{"Dockerfile", "FROM php:7"},
		{"gopher.txt", "Gopher names:\nGeorge\nGeoffrey\nGonzo"},
		{"todo.txt", "Get animal handling license."},
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
	buildCtx := bytes.NewReader(buf.Bytes())
	tr := tar.NewReader(buildCtx)

	// Iterate through the files in the archive.
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Fprintln(w, "File:", hdr.Name)
	}

	// Docker build
	buildOptions := types.ImageBuildOptions{
		Tags:           []string{"testtko"},
	}
	//response, err := dockercli.ImageBuild(context.Background(), buildCtx, buildOptions)
	dockercli.ImageBuild(context.Background(), buildCtx, buildOptions)
}
