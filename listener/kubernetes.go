package listener

import (
	"fmt"
	"log"
	"os"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/watch"
)

//A file of utilities to make working with K8s easier.

//GetClient returns a Kubernetes client.
func GetClient() (*client.Client, error) {
	var kubeConfig restclient.Config

	// Set the Kubernetes configuration based on the environment
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		config, err := restclient.InClusterConfig()

		if err != nil {
			return nil, fmt.Errorf("Failed to create in-cluster config: %v.", err)
		}

		kubeConfig = *config
	} else {
		kubeConfig = restclient.Config{
			Host: os.Getenv("KUBE_HOST"),
		}

		if kubeConfig.Host == "" {
			return nil, fmt.Errorf("You must specify the KUBE_HOST env var ")
		}
	}

	// Create the Kubernetes client based on the configuration
	return client.New(&kubeConfig)
}

//GetSecretWatcher create a secret watcher from the given namespace
func GetSecretWatcher(namespace string) watch.Interface {

	kubeClient, err := GetClient()

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	//only get secrets
	secretWatchOptions := api.ListOptions{
		Watch: true,
	}

	secretWatcher, err := kubeClient.Secrets(namespace).Watch(secretWatchOptions)

	if err != nil {
		log.Fatalf("Failed to create secret watcher: %v.", err)
	}

	return secretWatcher

}
