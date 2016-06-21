#Format is MAJOR . MINOR . PATCH

IMAGE_VERSION=0.1.1


test-build-and-package: test-source build-and-package

build-and-push-to-hub: build-and-package push-to-hub

build-and-package: compile-linux build-image


test-source:
	go test -v $$(glide novendor)

compile-linux:	
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w' -o build/secretrestart .
	
build-image:
	docker build -t thirtyx/secretrestart .

push-to-local:
	docker tag -f thirtyx/secretrestart localhost:5000/thirtyx/secretrestart
	docker push localhost:5000/thirtyx/secretrestart

push-to-hub:
	docker tag -f thirtyx/secretrestart thirtyx/secretrestart:$(IMAGE_VERSION)
	docker push thirtyx/secretrestart:$(IMAGE_VERSION)

deploy-to-kube:
	kubectl run secretrestart --image=localhost:5000/thirtyx/secretrestart:latest
