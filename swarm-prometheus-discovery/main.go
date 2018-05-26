package main

import (
	"encoding/json"
	"log"
	"sync"
	"github.com/fsouza/go-dockerclient"
	"fmt"
	"flag"
	"strings"
	"os"
	"github.com/Sirupsen/logrus"
	"time"
	"github.com/docker/docker/api/types/swarm"
)

var filters = make(map[string][]string)
var networkID = ""
var networkName = ""
var ignoredServices = make([]string, 0)

func init() {
	filters["desired-state"] = []string{"running"}
	networkName = *flag.String("network", "my_network", "Specify the name of the network you want to scrape")
	ignoredServicesStr := *flag.String("ignoredServices", "prometheus", "Comma-separated list of service names we do not want to scrape")
	// Stuff any services we don't want to scrape (such as ourselves and prometheus server)
	// into a slice.
	parseIgnoredServices(ignoredServicesStr)
}

func main() {
	logrus.Println("Starting Swarm-scraper!")

	// Connect to the Docker API
	endpoint := "unix:///var/run/docker.sock"
	dockerClient, err := docker.NewClient(endpoint)
	if err != nil {
		panic(err)
	}

	// Find the networkID we want to address tasks on.
	findNetworkId(dockerClient, networkName)

	// Start the task poller
	go func(dockerClient *docker.Client) {
		for {
			time.Sleep(time.Second * 15)
			pollTasks(dockerClient)
		}
	}(dockerClient)

	// Block...
	log.Println("Waiting at block...")

	wg := sync.WaitGroup{} // Use a WaitGroup to block main() exit
	wg.Add(1)
	wg.Wait()
}

func findNetworkId(dockerClient *docker.Client, networkName string) {
	networks, _ := dockerClient.ListNetworks()
	for _, val := range networks {
		if val.Name == networkName {
			networkID = val.ID
			return
		}
	}
	logrus.Errorf("Could not find NetworkID of %v, will assume 'ingress'\n", networkName)
	for _, val := range networks {
		if val.Name == "ingress" {
			networkID = val.ID
			return
		}
	}
	panic("Could neither resolve network " + networkName + " nor ingress network, panic!")
}

func parseIgnoredServices(ignoredServicesStr string) {
	if strings.Contains(ignoredServicesStr, ",") {
		copy(ignoredServices, strings.Split(ignoredServicesStr, ","))
	} else {
		ignoredServices = append(ignoredServices, ignoredServicesStr)
	}
	logrus.Printf("Ignored services: %v\n", ignoredServices)
}
func pollTasks(client *docker.Client) {

	tasks, _ := client.ListTasks(docker.ListTasksOptions{Filters: filters})
	tasksMap := make(map[string]*ScrapedTask)

	for _, task := range tasks {
		taskServiceId := task.ServiceID

		// Lookup service
		service, _ := client.InspectService(taskServiceId)

		// Skip if service is in ignoredList, e.g. don't scrape prometheus...
		if isInIgnoredList(service.Spec.Name) {
			continue
		}
		portNumber := "-1"

		// Find HTTP port?
		for _, port := range service.Endpoint.Ports {
			if port.Protocol == "tcp" {
				portNumber = fmt.Sprint(port.PublishedPort)
			}
		}

		// Skip if no exposed tcp port
		if portNumber == "-1" {
			continue
		}

		// Iterate network attachments on task
		for _, netw := range task.NetworksAttachments {

			// Only extract IP if on expected network.
			if netw.Network.ID == networkID {
				if taskEntry, ok := tasksMap[taskServiceId]; ok {
					processExistingTask(taskEntry, netw, portNumber, service)
				} else {
					processNewTask(netw, portNumber, service, tasksMap, taskServiceId)
				}
			}
		}
	}

	// Transform values of map into slice.
	taskList := make([]ScrapedTask, 0)
	for _, value := range tasksMap {
		taskList = append(taskList, *value)
	}

	// Write config file
	bytes, err := json.Marshal(taskList)
	if err != nil {
		panic(err)
	}

	file, err := os.Create("/etc/swarm-endpoints/swarm-endpoints.json")
	if err != nil {
		fmt.Errorf("Error writing file: %v\n", err.Error())
		panic(err.Error())
	}
	file.Write(bytes)
	file.Close()
}

func processNewTask(netw swarm.NetworkAttachment, portNumber string, service *swarm.Service, tasksMap map[string]*ScrapedTask, taskServiceId string) {
	// New task
	taskEntry := ScrapedTask{Targets: make([]string, 0), Labels: make(map[string]string)}
	for _, adr := range netw.Addresses {
		taskEntry.Targets = append(taskEntry.Targets, formatIp(adr, portNumber))
	}
	taskEntry.Labels["task"] = service.Spec.Name
	tasksMap[taskServiceId] = &taskEntry
}

func processExistingTask(taskEntry *ScrapedTask, netw swarm.NetworkAttachment, portNumber string, service *swarm.Service) {
	// Existing task
	localTargets := make([]string, len(taskEntry.Targets))
	copy(localTargets, taskEntry.Targets)
	for _, adr := range netw.Addresses {
		localTargets = append(localTargets, formatIp(adr, portNumber))
	}
	taskEntry.Targets = localTargets
	taskEntry.Labels["task"] = service.Spec.Name
}

func isInIgnoredList(s string) bool {
	for _, ignored := range ignoredServices {
		if ignored == s {
			return true
		}
	}
	return false
}
func formatIp(ip string, port string) string {
	// Remove /NN part of ip
	index := strings.Index(ip, "/")
	ip = ip[:index] + ":" + port
	return ip
}

type ScrapedTask struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}
