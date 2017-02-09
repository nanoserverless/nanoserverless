package main

import (
	"archive/tar"
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gorilla/mux"
	//"io"
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
	r.HandleFunc("/{base}/{name}", infofunc)
	r.HandleFunc("/{base}/{name}/create", create)
	r.HandleFunc("/{base}/{name}/exec", exec)
	r.HandleFunc("/{base}/{name}/up", up)
	r.HandleFunc("/{base}/{name}/down", down)
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

func infofunc(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	base := vars["base"]
	name := vars["name"]
	fmt.Fprintln(w, "You're trying to get info on the", base, name, "function")
	fmt.Fprintln(w, "But it's not implement yet :D")
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
				Image: tag,
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

func exec(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	base := vars["base"]
	name := vars["name"]
	tag := tagprefix + "-" + base + "-" + name
	containername := tag
	ctx := context.Background()

	/*cmd := []string{}
	if base == "php7" {
		cmd = []string{"php", "app"}
	}
	if base == "node7" {
		cmd = []string{"node", "app"}
	}*/

	// Test if we can http to the container
	resp_http, err := http.Get("http://" + containername)
	if err != nil {
		// Create
		resp, err := dockercli.ContainerCreate(ctx, &container.Config{
			Image:      tag,
			Entrypoint: []string{"/run"},
			//      AttachStdout: true,
		}, nil, nil, containername)
		if err != nil {
			panic(err)
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
		responseBody, err := dockercli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
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
		stdcopy.StdCopy(w, w, responseBody)

		// Delete
		_ = dockercli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})

		panic(err)
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

	// Generate dockerfile
	dockerfile := "FROM "
	if base == "php7" {
		dockerfile += "php:7"
	}
	if base == "node7" {
		dockerfile += "node:7"
	}
	dockerfile += "\nCOPY shell2http /"
	dockerfile += "\nCOPY app /"
	dockerfile += "\nCOPY run /"
	if base == "php7" {
		dockerfile += "\nENTRYPOINT [\"/shell2http\", \"-port=80\", \"-cgi\", \"-export-all-vars\", \"/\", \"/run\"]"
	}
	if base == "node7" {
		dockerfile += "\nENTRYPOINT [\"/shell2http\", \"-port=80\", \"-cgi\", \"-export-all-vars\", \"/\", \"/run\"]"
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

	// Generate run
	run := ""
	if base == "php7" {
		run += "php app"
	}
	if base == "node7" {
		run += "node app"
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
		Tags:           []string{tag},
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
