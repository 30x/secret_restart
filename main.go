package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"k8s.io/kubernetes/pkg/api"

	"strconv"

	"k8s.io/kubernetes/pkg/watch"

	"github.com/30x/k8s-pods-ingress/kubernetes"
	"github.com/30x/secretrestart/listener"
)

// Boostrap script that performs sanity checking to ensure our pods exist
func main() {

	//check our env and die if not set
	secretName := os.Getenv("SECRET_NAME")

	if secretName == "" {
		log.Fatal("You must specifiy the SECRET_NAME environment variable")
	}

	ignoreCountString := os.Getenv("IGNORE_COUNT")

	if ignoreCountString == "" {
		log.Fatal("You must specifiy the IGNORE_COUNT environment variable")
	}

	ignoreCount, err := strconv.Atoi(ignoreCountString)

	if err != nil {
		log.Fatal("You must specify a valid integer for the IGNORE_COUNT value")
	}

	podNamespace := os.Getenv("POD_NAMESPACE")

	if podNamespace == "" {
		log.Fatal("You must specifiy the POD_NAMESPACE environment variable")
	}

	podName := os.Getenv("POD_NAME")

	if podName == "" {
		log.Fatal("You must specifiy the POD_NAME environment variable")
	}

	log.Print("Connecting to kubernetes")

	client, err := kubernetes.GetClient()

	if err != nil {
		log.Fatalf("Unable to connect to kuberentes existing: %v", err)
	}

	emitter := listener.CreateK8sSecretEmitter(podNamespace)

	seed := time.Now().Unix()
	rand.Seed(seed)

	log.Print("Started watcher")

	channel := emitter.Watcher().ResultChan()

	eventReceivedCount := 0

	for {

		//TODO, do we need to worry about re-connect?
		event, ok := <-channel

		if !ok {
			log.Fatal("Kubernetes closed the secret watcher, existing")
		}

		log.Printf("Received event %v", event)

		secret := event.Object.(*api.Secret)

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

		//continue, then do our random shutdown
		eventReceivedCount++

		if eventReceivedCount <= ignoreCount {
			log.Printf("Received a valid event.  Received count is %d and ignoreCount is %d, ignoring", eventReceivedCount, ignoreCount)

			continue
		}

		//add a 10 minute variance to our shutdown timer
		waitTime := rand.Intn(60)

		log.Printf("Shutting down pod in %d seconds", waitTime)

		//set the timer and wait
		timer := time.NewTimer(time.Duration(waitTime) * time.Second)
		<-timer.C

		var gracePeriod int64

		gracePeriod = 0

		log.Printf("Shutting down pod %s in namespace %s", podName, podNamespace)

		client.Pods(podNamespace).Delete(podName, &api.DeleteOptions{GracePeriodSeconds: &gracePeriod})

		//we deliberately let this loop back to the top.  If we're a sidecar, this will terminate when the pod terminates

	}
}
