package listener

import (
	"k8s.io/kubernetes/pkg/watch"
)

type k8sSecretEmitter struct {
	watcher watch.Interface
}

//CreateK8sSecretEmitter create the secret emitter for the k8s system
func CreateK8sSecretEmitter(namespace string) SecretEmitter {

	watcher := GetSecretWatcher(namespace)

	return &k8sSecretEmitter{watcher: watcher}
}

//Channel Get the channel that emits watch.Event with api.Secret objects only
func (k8sSecretEmitter *k8sSecretEmitter) Watcher() watch.Interface {

	return k8sSecretEmitter.watcher
}
