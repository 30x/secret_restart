package listener

import "k8s.io/kubernetes/pkg/watch"

//SecretEmitter listens and emits events on the channel passed
type SecretEmitter interface {

	//Channel Get the channel that emits watch.Event with api.Secret objects only
	Watcher() watch.Interface
}
