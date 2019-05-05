package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

// target represent a target scrapable by promtail
type target struct {
	Targets []string     `json:"targets"`
	Labels  TargetLabels `json:"Labels,omitempty"`
}

// containers represents a subset of container specs
type containers []struct {
	ID     string          `json:"Id"`
	Labels ContainerLabels `json:"Labels,omitempty"`
}

// ContainerLabels represents labels we are extracting
type ContainerLabels struct {
	App                       string `json:"app"`
	Application               string `json:"application"`
	Component                 string `json:"component"`
	Stack                     string `json:"com.docker.stack.namespace"`
	ComDockerSwarmServiceName string `json:"com.docker.swarm.service.name"`
	Type                      string `json:"type"`
}

// TargetLabels represents labels we are writing
type TargetLabels struct {
	App                       string `json:"app,omitempty"`
	Application               string `json:"application,omitempty"`
	Component                 string `json:"component,omitempty"`
	Stack                     string `json:"stack,omitempty"`
	ComDockerSwarmServiceName string `json:"service,omitempty"`
	Type                      string `json:"type,omitempty"`
}

func main() {

	every := flag.Duration("every", 10*time.Second, "How often to refresh the targets.")
	where := flag.String("out", "/etc/promtail/promtail-targets.json", "Path where the target file will be written.")
	what := flag.String("url", "/var/run/docker.sock", "Docker socket path.")

	flag.Parse()

	tick := time.Tick(*every)
	fmt.Println("Starting run loop")
	for {
		select {
		case <-tick:
			containersToTargets(*what, *where)
		}
	}
}

func containersToTargets(url, to string) {

	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", url)
			},
		},
	}

	var response *http.Response
	var err error

	response, err = httpc.Get("http://unix/v1.24/containers/json")
	if err != nil {
		fmt.Printf("Error while getting containers: %s\n", err)
		return
	}

	defer response.Body.Close() //nolint

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error reading body: %s\n", err)
		return
	}

	var containers containers
	if err = json.Unmarshal(body, &containers); err != nil {
		fmt.Printf("Error unmarshalling containers: %s\n", err)
		return
	}

	targets := []target{}
	for _, c := range containers {
		targets = append(targets, target{
			Targets: []string{c.ID},
			Labels:  TargetLabels(c.Labels),
		})
	}

	file, err := json.MarshalIndent(targets, "", " ")
	if err != nil {
		fmt.Printf("Error marshalling targets: %s\n", err)
		return
	}

	err = ioutil.WriteFile(to, file, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %s\n", err)
		return
	}

}
