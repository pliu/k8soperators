APP_NAME=k8soperators
CLUSTER_NAME=k8soperators
KIND_IMAGE=kindest/node:v1.16.4@sha256:b91a2c2317a000f3a783489dfb755064177dbc3a0b2f4147d50f04825d016f55
VERSION=$$(grep -o '".*"' version/version.go | sed 's/"//g')
APP_IMAGE=$(APP_NAME):$(VERSION)

.PHONY: kind_create
kind_create:
	kind create cluster --config=kind/config.yaml --name $(CLUSTER_NAME) --image $(KIND_IMAGE)

.PHONY: kind_destroy
kind_destroy:
	kind delete cluster --name $(CLUSTER_NAME)

.PHONY: build
build:
	operator-sdk build $(APP_IMAGE)

.PHONY: template_deployment
template_deployment:
	sed "s|REPLACE_IMAGE|$(APP_IMAGE)|g" deploy-templates/deployment.yaml > deploy/deployment.yaml

.PHONY: integration_test
integration_test: build template_deployment
	-make kind_create
	kind load docker-image $(APP_IMAGE) --name $(CLUSTER_NAME)
	kubectl apply -f ./deploy
	kubectl delete -f ./deploy

.PHONY: unit_test
unit_test:
	go test ./...

.PHONY: run_locally
run_locally:
	OPERATOR_NAME=$(APP_NAME) operator-sdk run --local
