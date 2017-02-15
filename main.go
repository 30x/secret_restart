package main

import (
	"log"
	"math/big"
	"math/rand"
	"net"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/rest"

	"fmt"
	"strconv"
)

// Boostrap script that performs sanity checking to ensure our pods exist
func main() {
	//check our env and die if not set
	secretName := os.Getenv("SECRET_NAME")

	if secretName == "" {
		log.Fatal("You must specifiy the SECRET_NAME environment variable")
	}

	shutdownTimespanString := os.Getenv("SHUTDOWN_TIMESPAN")

	if shutdownTimespanString == "" {
		log.Fatal("You must specifiy the SHUTDOWN_TIMESPAN environment variable")
	}

	shutdownTimespan, err := strconv.Atoi(shutdownTimespanString)

	if err != nil {
		log.Fatal("You must specify a valid integer for the SHUTDOWN_TIMESPAN value")
	}

	podNamespace := os.Getenv("POD_NAMESPACE")

	if podNamespace == "" {
		log.Fatal("You must specifiy the POD_NAMESPACE environment variable")
	}

	podName := os.Getenv("POD_NAME")

	if podName == "" {
		log.Fatal("You must specifiy the POD_NAME environment variable")
	}

	podIP := os.Getenv("POD_IP")

	if podIP == "" {
		log.Fatal("You must specifiy the POD_IP environment variable")
	}

	log.Print("Connecting to kubernetes")

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	client, err := kubernetes.NewForConfig(config)

	if err != nil {
		log.Fatalf("Unable to connect to kubernetes exiting: %v", err)
	}

	watcher, err := getWatcher(podNamespace, client)

	if err != nil {
		log.Printf("Unable to get a watcher on first try: %v", err)
		log.Println("Trying again to get a watcher.")

		watcher, err = getWatcher(podNamespace, client)
		if err != nil {
			log.Fatalf("Tryed to get a watcher again, and it failed, exiting: %v", err)
		}
	}

	ipvSeed := ip4toInt(podIP)

	rand.Seed(ipvSeed)

	log.Print("Started watcher")

	channel := watcher.ResultChan()

	for {
		event, ok := <-channel

		if !ok { // check this first so that if the first time failed, we try getting a watcher again first
			log.Print("Kubernetes watcher was closed, recreating watcher.")

			watcher, err = getWatcher(podNamespace, client)

			if err != nil {
				log.Fatalf("Unable to get a watcher on restart, exiting: %v", err)
			}

			channel = watcher.ResultChan()
		} else {
			log.Printf("Received event %s", event.Type)

			secret := event.Object.(*v1.Secret)

			// Only record secret events for secrets with the name we are interested in
			if secret.Name != secretName {
				log.Printf("Received event for secret '%s'. Need secret name '%s', ignoring", secret.Name, secretName)
				continue
			}

			//if it's not a type we care about, ignore
			if event.Type != watch.Added && event.Type != watch.Modified {
				log.Printf("Received event for secret '%s' of type %s.  Ignoring", secretName, event.Type)
				continue
			}

			log.Printf("Received event for secret %s, shutting down", secretName)

			//add a 10 minute variance to our shutdown timer
			waitTime := rand.Intn(shutdownTimespan)

			log.Printf("Shutting down pod in %d seconds", waitTime)

			//set the timer and wait
			timer := time.NewTimer(time.Duration(waitTime) * time.Second)
			<-timer.C

			var gracePeriod int64

			log.Printf("Shutting down pod %s in namespace %s", podName, podNamespace)
			client.Pods(podNamespace).Delete(podName, &v1.DeleteOptions{GracePeriodSeconds: &gracePeriod})
		}
	}
}

func ip4toInt(ipv4Ip string) int64 {

	ipv4Address := net.ParseIP(ipv4Ip)

	IPv4Int := big.NewInt(0)
	IPv4Int.SetBytes(ipv4Address.To4())
	return IPv4Int.Int64()
}

func getWatcher(namespace string, client *kubernetes.Clientset) (watch.Interface, error) {
	existingSecrets, err := client.Secrets(namespace).List(v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Unable to get list for most recent resource version: %v", err)
	}

	return client.Secrets(namespace).Watch(v1.ListOptions{
		ResourceVersion: existingSecrets.ResourceVersion,
	})
}
