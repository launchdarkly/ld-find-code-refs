IMAGE_NAME=$(shell basename $(CURDIR))-flag-parser

.PHONY: linux-compile
.linux-compile:
	GOOS=linux GOARCH=amd64 go build -o out/$(IMAGE_NAME)

.docker-login:
	docker login

.PHONY: docker-build
.docker-build: linux-compile .docker-login
	test $(TAG) || (echo "Please provide tag"; exit 1)
	docker build -t $(IMAGE_NAME):$(TAG) .

.PHONY: docker-tag
.docker-tag:
	test $(DOCKER_REPO) || (echo "Please provide DOCKER_REPO"; exit 1)
	test $(TAG) || (echo "Please provide tag"; exit 1)
	docker tag $(IMAGE_NAME):$(TAG) $(DOCKER_REPO)/$(IMAGE_NAME):$(TAG)

.PHONY: docker-push
.docker-push:
	test $(DOCKER_REPO) || (echo "Please provide DOCKER_REPO"; exit 1)
	test $(TAG) || (echo "Please provide TAG"; exit 1)
	docker push $(DOCKER_REPO)/$(IMAGE_NAME):$(TAG)

.PHONY: all
all: .linux-compile .docker-build .docker-tag .docker-push

