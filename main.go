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
        Name    []string          `json:"Names"`
        Image string `json:"Image"`
}

// ContainerLabels represents labels we are extracting
type ContainerLabels struct {
        Ecs_cluster                       string `json:"com.amazonaws.ecs.cluster,omitempty"`
        Ecs_task_definition_family      string `json:"com.amazonaws.ecs.task-definition-family,omitempty"`
        Ecs_task_definition_version      string `json:"com.amazonaws.ecs.task-definition-version,omitempty"`
        Ecs_task_definition_container_name      string `json:"com.amazonaws.ecs.container-name,omitempty"`
        Ecs_task_arn      string `json:"com.amazonaws.ecs.task-arn,omitempty"`
        DockerComposeService  string `json:"com.docker.compose.service,omitempty"`
        Docker_Composeproject  string `json:"com.docker.compose.project,omitempty"`
        Image string
        Name string
}

// TargetLabels represents labels we are writing
type TargetLabels struct {
        Ecs_cluster                       string `json:"ecs_cluster,omitempty"`
        Ecs_task_definition_family      string `json:"ecs_task_definition_family,omitempty"`
        Ecs_task_definition_version      string `json:"ecs_task_definition_version,omitempty"`
        Ecs_task_definition_container_name      string `json:"ecs_container_name,omitempty"`
        Ecs_task_arn      string `json:"ecs_task_arn,omitempty"`
        DockerComposeService  string `json:"docker_compose_service,omitempty"`
        Docker_Composeproject  string `json:"docker_compose_project,omitempty"`
        Image string
        Name string
}

var globalTransport *http.Transport


func main() {

        every := flag.Duration("every", 10*time.Second, "How often to refresh the targets.")
        where := flag.String("out", "/etc/promtail/promtail-targets.json", "Path where the target file will be written.")
        what := flag.String("url", "/var/run/docker.sock", "Docker socket path.")

        flag.Parse()

        init_http(*what)
        tick := time.Tick(*every)
        fmt.Println("Starting run loop")
        for {
                select {
                case <-tick:
                        containersToTargets(*what, *where)
                }
        }
}

func init_http(url string) {
        globalTransport = &http.Transport{
            DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
                return net.Dial("unix", url)
            },
        }
}
func containersToTargets(url, to string) {

        httpc := http.Client{
                Transport: globalTransport,
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
                if c.Name[0] != "/ecs-agent" {
                        target_labels := TargetLabels(c.Labels)
                        target_labels.Name = c.Name[0]
                        target_labels.Image = c.Image
                        targets = append(targets, target{
                                Targets: []string{c.ID},
                                Labels:  target_labels,
                        })
                }
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
