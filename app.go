package main

import (
	"archive/tar"
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gorilla/mux"
	//"io"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var port string
var dockercli, err = client.NewClient("unix:///var/run/docker.sock", "v1.25", nil, map[string]string{"User-Agent": "nanoserverless"})
var tagprefix = "nanoserverless"
var registry string

type base struct {
	Run        string
	ViewCode   []string
	FromImg    string
	ExtraBuild string
}

var bases = make(map[string]base)

func init() {
	flag.StringVar(&port, "port", "80", "give me a port number")
	registry = os.Getenv("REGISTRY_URL")
	if registry != "" {
		registry += "/"
		synchroRepo()
	}

	bases["php7"] = base{
		"#!/bin/sh\nphp app",
		[]string{"cat", "/app"},
		"php:7",
		"",
	}
	bases["node7"] = base{
		"#!/bin/sh\nnode app",
		[]string{"cat", "/app"},
		"node:7",
		"",
	}
	bases["java8"] = base{
		"#!/bin/sh\njava app",
		[]string{"cat", "/app.java"},
		"openjdk:8",
		"RUN mv app app.java && javac app.java",
	}
	bases["go17"] = base{
		"",
		[]string{"cat", "/go/src/app.go"},
		"golang:1.7",
		"WORKDIR /go/src\nENV CGO_ENABLED=0\nENV GO_PATH=/go/src\nRUN mv /app ./app.go && go build -a --installsuffix cgo --ldflags=-s -o /run",
	}
	bases["python27"] = base{
		"#!/bin/sh\npython app",
		[]string{"cat", "/app"},
		"python:2.7",
		"",
	}
}

func main() {
	// defer profile.Start().Stop()
	flag.Parse()

	// Router
	r := mux.NewRouter()
	r.HandleFunc("/list", list)
	r.HandleFunc("/{base}/{name}", infofunc)
	r.HandleFunc("/{base}/{name}/create", create)
	r.HandleFunc("/{base}/{name}/exec", exec)
	r.HandleFunc("/{base}/{name}/up", up)
	r.HandleFunc("/{base}/{name}/down", down)
	r.HandleFunc("/{base}/{name}/code", code)
	r.HandleFunc("/whoami", whoami)
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

func synchroRepo() {
	ctx := context.Background()
	resp_http, err := http.Get("http://" + registry + "v2/_catalog")
	if err != nil {
		log.Fatal(err)
	}
	defer resp_http.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp_http.Body)
	//fmt.Fprint(w, buf.String())
	type Repos struct {
		Repositories []string `json:"repositories"`
	}
	var repos Repos
	err = json.Unmarshal(buf.Bytes(), &repos)
	if err != nil {
		fmt.Println("error:", err)
	}
	for _, tag := range repos.Repositories {
		// Pull all start by "tagprefix-"
		if strings.HasPrefix(tag, tagprefix+"-") {
			fmt.Println("Pulling image :", registry+tag)
			// Trying to pull image
			resp_pull, err := dockercli.ImagePull(ctx, registry+tag, types.ImagePullOptions{
				RegistryAuth: "ewogICJ1c2VybmFtZSI6ICIiLAogICJwYXNzd29yZCI6ICIiLAogICJlbWFpbCI6ICIiLAogICJzZXJ2ZXJhZGRyZXNzIjogIiIKfQo=",
			})
			// Wait pull finish
			buf_pull := new(bytes.Buffer)
			buf_pull.ReadFrom(resp_pull)
			buf_pull.String()
			if err != nil {
				panic(err)
			}
		}
	}

}

func list(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()

	// Get images from registry
	if registry != "" {
		synchroRepo()
	}

	// List local images
	images, err := dockercli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	for _, image := range images {
		for _, tag := range image.RepoTags {
			if strings.HasPrefix(tag, registry+tagprefix+"-") {
				fmt.Fprintln(w, "local tag", tag)
			}
		}
	}
}

func infofunc(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	base := vars["base"]
	name := vars["name"]
	tag := tagprefix + "-" + base + "-" + name
	servicename := tag
	ctx := context.Background()

	fmt.Fprintln(w, "Service", servicename, "status :")

	// Get tasks
	serviceNameFilter := filters.NewArgs()
	serviceNameFilter.Add("name", servicename)
	tasks, err := dockercli.TaskList(ctx, types.TaskListOptions{
		Filters: serviceNameFilter,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, task := range tasks {
		fmt.Fprintln(w, "Task", task.Slot, task.Status.ContainerStatus.ContainerID, task.Status.State, "("+task.Status.Message+")")
	}
	if len(tasks) == 0 {
		fmt.Fprintln(w, "Not UP")
	}

	//fmt.Fprintln(w, "You're trying to get info on the", base, name, "function")
	//fmt.Fprintln(w, "But it's not implement yet :D")
}

func down(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	base := vars["base"]
	name := vars["name"]
	tag := tagprefix + "-" + base + "-" + name
	servicename := tag
	ctx := context.Background()

	err := dockercli.ServiceRemove(ctx, servicename)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(w, "Service", servicename, "removed")

}

func up(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	base := vars["base"]
	name := vars["name"]
	tag := tagprefix + "-" + base + "-" + name
	servicename := tag
	ctx := context.Background()

	/* Goal :
	docker service create \
	  --name nanoserverless-node7-pi \
	  --network nanoserverless \
	  nanoserverless-node7-pi
	*/

	// Network
	network := swarm.NetworkAttachmentConfig{
		Target: "nanoserverless",
	}

	// Create
	service := swarm.ServiceSpec{
		Annotations: swarm.Annotations{
			Name: servicename,
			//Labels: runconfigopts.ConvertKVStringsToMap(opts.labels.GetAll()),
		},
		TaskTemplate: swarm.TaskSpec{
			ContainerSpec: swarm.ContainerSpec{
				Image: registry + tag,
				/*Args:     opts.args,
				Env:      currentEnv,
				Hostname: opts.hostname,
				Labels:   runconfigopts.ConvertKVStringsToMap(opts.containerLabels.GetAll()),
				Dir:      opts.workdir,
				User:     opts.user,
				Groups:   opts.groups.GetAll(),
				TTY:      opts.tty,
				ReadOnly: opts.readOnly,
				Mounts:   opts.mounts.Value(),
				DNSConfig: &swarm.DNSConfig{
					Nameservers: opts.dns.GetAll(),
					Search:      opts.dnsSearch.GetAll(),
					Options:     opts.dnsOption.GetAll(),
				},
				Hosts:           convertExtraHostsToSwarmHosts(opts.hosts.GetAll()),
				StopGracePeriod: opts.stopGrace.Value(),
				Secrets:         nil,
				Healthcheck:     healthConfig,*/
			},
			Networks: []swarm.NetworkAttachmentConfig{network},
			/*Resources:     opts.resources.ToResourceRequirements(),
			RestartPolicy: opts.restartPolicy.ToRestartPolicy(),
			Placement: &swarm.Placement{
				Constraints: opts.constraints.GetAll(),
			},
			LogDriver: opts.logDriver.toLogDriver(),*/
		},
		//Networks: convertNetworks(opts.networks.GetAll()),
		/*Mode:     serviceMode,
		UpdateConfig: &swarm.UpdateConfig{
			Parallelism:     opts.update.parallelism,
			Delay:           opts.update.delay,
			Monitor:         opts.update.monitor,
			FailureAction:   opts.update.onFailure,
			MaxFailureRatio: opts.update.maxFailureRatio.Value(),
		},
		EndpointSpec: opts.endpoint.ToEndpointSpec(),*/
	}

	resp, err := dockercli.ServiceCreate(ctx, service, types.ServiceCreateOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(w, "Service id ", resp.ID, "created")
}

func code(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	base := vars["base"]
	name := vars["name"]
	tag := tagprefix + "-" + base + "-" + name
	ctx := context.Background()

	baseStruct, ok := bases[base]
	if !ok {
		fmt.Fprintln(w, base, "not supported yet !")
		return
	}

	// Create
	resp, err := dockercli.ContainerCreate(ctx, &container.Config{
		Image:      registry + tag,
		Entrypoint: baseStruct.ViewCode,
		//      AttachStdout: true,
	}, nil, nil, "")
	if err != nil {
		// Trying to pull image
		resp_pull, err := dockercli.ImagePull(ctx, registry+tag, types.ImagePullOptions{
			RegistryAuth: "ewogICJ1c2VybmFtZSI6ICIiLAogICJwYXNzd29yZCI6ICIiLAogICJlbWFpbCI6ICIiLAogICJzZXJ2ZXJhZGRyZXNzIjogIiIKfQo=",
		})
		// Wait pull finish
		buf_pull := new(bytes.Buffer)
		buf_pull.ReadFrom(resp_pull)
		buf_pull.String()

		if err != nil {
			panic(err)
		}

		resp, err = dockercli.ContainerCreate(ctx, &container.Config{
			Image:      registry + tag,
			Entrypoint: baseStruct.ViewCode,
		}, nil, nil, "")
		if err != nil {
			panic(err)
		}
	}

	// Run
	if err := dockercli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	// Wait
	if _, err = dockercli.ContainerWait(ctx, resp.ID); err != nil {
		panic(err)
	}

	// Logs
	responseBody, err := dockercli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		panic(err)
	}
	defer responseBody.Close()

	// Print
	/*buf := new(bytes.Buffer)
	  buf.ReadFrom(out)
	  result := buf.String()
	  fmt.Fprintln(w, result)*/

	//io.Copy(w, []byte(out))
	/*fmt.Fprintln(w, "Result:")
	  buf := new(bytes.Buffer)
	  buf.ReadFrom(responseBody)
	  newStr := buf.String()*/

	stdcopy.StdCopy(w, w, responseBody)
	//fmt.Fprintln(w, newStr)

	// Delete
	_ = dockercli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})

}

func exec(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	base := vars["base"]
	name := vars["name"]
	tag := tagprefix + "-" + base + "-" + name
	servicename := tag
	ctx := context.Background()

	// Test if we can http to the service
	resp_http, err := http.Get("http://" + servicename)
	if err != nil {
		// Create
		resp, err := dockercli.ContainerCreate(ctx, &container.Config{
			Image:      registry + tag,
			Entrypoint: []string{"/run"},
			//      AttachStdout: true,
		}, nil, nil, "")
		if err != nil {
			// Trying to pull image
			resp_pull, err := dockercli.ImagePull(ctx, registry+tag, types.ImagePullOptions{
				RegistryAuth: "ewogICJ1c2VybmFtZSI6ICIiLAogICJwYXNzd29yZCI6ICIiLAogICJlbWFpbCI6ICIiLAogICJzZXJ2ZXJhZGRyZXNzIjogIiIKfQo=",
			})
			// Wait pull finish
			buf_pull := new(bytes.Buffer)
			buf_pull.ReadFrom(resp_pull)
			buf_pull.String()

			if err != nil {
				panic(err)
			}

			resp, err = dockercli.ContainerCreate(ctx, &container.Config{
				Image:      registry + tag,
				Entrypoint: []string{"/run"},
			}, nil, nil, "")
			if err != nil {
				panic(err)
			}
		}

		// Run
		if err := dockercli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
			panic(err)
		}

		// Wait
		if _, err = dockercli.ContainerWait(ctx, resp.ID); err != nil {
			panic(err)
		}

		// Logs
		responseBody, err := dockercli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
		if err != nil {
			panic(err)
		}
		defer responseBody.Close()

		// Print
		/*buf := new(bytes.Buffer)
		  buf.ReadFrom(out)
		  result := buf.String()
		  fmt.Fprintln(w, result)*/

		//io.Copy(w, []byte(out))
		/*fmt.Fprintln(w, "Result:")
		buf := new(bytes.Buffer)
		buf.ReadFrom(responseBody)
		newStr := buf.String()*/

		stdcopy.StdCopy(w, w, responseBody)
		//fmt.Fprintln(w, newStr)

		// Delete
		_ = dockercli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})

	} else {
		defer resp_http.Body.Close()
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp_http.Body)
		fmt.Fprint(w, buf.String())
	}

}

func create(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	base := vars["base"]
	name := vars["name"]
	tag := tagprefix + "-" + base + "-" + name
	bodyb, _ := ioutil.ReadAll(req.Body)
	body := string(bodyb)

	baseStruct, ok := bases[base]
	if !ok {
		fmt.Fprintln(w, base, "not supported yet !")
		return
	}

	// Generate dockerfile
	dockerfile := "FROM "
	dockerfile += baseStruct.FromImg
	dockerfile += "\nCOPY shell2http /"
	dockerfile += "\nCOPY app /"
	dockerfile += "\nCOPY run /"
	dockerfile += "\n" + baseStruct.ExtraBuild
	dockerfile += "\nENTRYPOINT [\"/shell2http\", \"-port=80\", \"-cgi\", \"-export-all-vars\", \"/\", \"/run\"]"

	// Generate app
	//app := ""
	//app += baseStruct.PreCode + "\n"
	app := body
	//app += baseStruct.PostCode

	// Generate run
	run := baseStruct.Run

	// Buffer context
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	// Add some files
	var files = []struct {
		Name, Body string
	}{
		{"Dockerfile", dockerfile},
		{"app", app},
		{"run", run},
	}
	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0700,
			Size: int64(len(file.Body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			log.Fatalln(err)
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			log.Fatalln(err)
		}
	}

	// Add shell2http
	dat, err := ioutil.ReadFile("/shell2http")
	if err != nil {
		log.Fatal(err)
	}
	hdr := &tar.Header{
		Name: "/shell2http",
		Mode: 0700,
		Size: int64(len(dat)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		log.Fatalln(err)
	}
	if _, err := tw.Write(dat); err != nil {
		log.Fatalln(err)
	}

	// Make sure to check the error on Close.
	if err := tw.Close(); err != nil {
		log.Fatalln(err)
	}

	// Open the tar archive for reading.
	reader := bytes.NewReader(buf.Bytes())
	// Docker build
	buildOptions := types.ImageBuildOptions{
		Tags:           []string{registry + tag},
		NoCache:        true,
		SuppressOutput: true,
	}

	response, err := dockercli.ImageBuild(context.Background(), reader, buildOptions)
	if err != nil {
		log.Fatalln(err)
		//fmt.Fprintln(w, "Error in creating image", tag)
	}
	defer response.Body.Close()
	buf2 := new(bytes.Buffer)
	buf2.ReadFrom(response.Body)
	result := buf2.String()
	fmt.Fprintln(w, "Image ", registry+tag, "created !\n")
	fmt.Fprintln(w, "Dockerfile:")
	fmt.Fprintln(w, dockerfile, "\n")
	fmt.Fprintln(w, "Code:")
	fmt.Fprintln(w, app, "\n")
	fmt.Fprintln(w, "Log:")
	fmt.Fprintln(w, result, "\n")
	//fmt.Fprintln(w, "response:", response.Body)
	//buildCtx := ioutil.NopCloser(reader)
	//dockercli.ImageBuild(context.Background(), buildCtx, buildOptions)

	// Push image if registry
	if registry != "" {
		response_push, err := dockercli.ImagePush(context.Background(), registry+tag, types.ImagePushOptions{
			RegistryAuth: "ewogICJ1c2VybmFtZSI6ICIiLAogICJwYXNzd29yZCI6ICIiLAogICJlbWFpbCI6ICIiLAogICJzZXJ2ZXJhZGRyZXNzIjogIiIKfQo=",
		})
		if err != nil {
			log.Fatalln(err)
		}
		buf3 := new(bytes.Buffer)
		buf3.ReadFrom(response_push)
		result_push := buf3.String()
		fmt.Fprintln(w, "Push:")
		fmt.Fprintln(w, result_push, "\n")
	}
}
