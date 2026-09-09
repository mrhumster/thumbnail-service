IMAGE_NAME := xomrkob/thumbnail-service
NAMESPACE := go-app
DEPLOYMENT := thumbnail-service
VERSION ?= $(shell git describe --tags --always || echo "latest")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
MODULE_NAME := github.com/mrhumster/thumbnail-service

PROTO_DIR := proto/stream
GEN_DIR := gen/go

.PHONY: build test vet proto docker-build docker-push docker-deploy keda-deploy apply-keda logs

build:
	go build -o thumbnail-worker ./cmd/worker/main.go

test:
	go test -v ./...

vet:
	go vet ./...

proto:
	@echo "Generate protoc"
	mkdir -p $(GEN_DIR)
	protoc --proto_path=$(PROTO_DIR) \
		--go_out=. --go-grpc_out=. \
		--go_opt=module=$(MODULE_NAME) \
		--go-grpc_opt=module=$(MODULE_NAME) \
		$(PROTO_DIR)/*.proto
	@echo "Proto file generated in $(GEN_DIR)"

docker-build:
	@echo "Building docker image $(IMAGE_NAME):$(VERSION)..."
	docker build -f Dockerfile \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(IMAGE_NAME):$(VERSION) \
		-t $(IMAGE_NAME):latest ..

docker-push:
	@echo "Pushing image $(IMAGE_NAME):$(VERSION)..."
	docker push $(IMAGE_NAME):$(VERSION)
	docker push $(IMAGE_NAME):latest

docker-deploy:
	@echo "Updating K8s deployment..."
	kubectl -n $(NAMESPACE) set image deployment/$(DEPLOYMENT) \
		$(DEPLOYMENT)=$(IMAGE_NAME):$(VERSION)
	@echo "Success!"

keda-deploy:
	@echo "Deploy KEDA"
	@echo "Add helm repository"
	helm repo add kedacore https://kedacore.github.io/charts
	@echo "Update repo..."
	helm repo update
	@echo "Install..."
	helm upgrade --install keda kedacore/keda --namespace keda --create-namespace
	@echo "Wait...."
	kubectl wait --for=condition=Ready pod --all -n keda --timeout=90s

apply-keda:
	@echo "Apply KEDA ScaledObject"
	kubectl apply -f deploy/k8s/keda/.

logs:
	kubectl -n $(NAMESPACE) logs -f -l app=$(DEPLOYMENT)
