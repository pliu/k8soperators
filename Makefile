CLUSTER_NAME=k8soperators
KIND_IMAGE=kindest/node:v1.16.4@sha256:b91a2c2317a000f3a783489dfb755064177dbc3a0b2f4147d50f04825d016f55
APP_NAME=k8soperators
VERSION=$$(grep -o '".*"' version/version.go | sed 's/"//g')
APP_IMAGE=$(APP_NAME):$(VERSION)
TESTING_NAMESPACE=integration-test

.PHONY: kind_create
kind_create:
	kind create cluster --config=kind/config.yaml --name $(CLUSTER_NAME) --image $(KIND_IMAGE)

.PHONY: kind_destroy
kind_destroy:
	kind delete cluster --name $(CLUSTER_NAME)

.PHONY: generate_code
generate_code:
	operator-sdk generate k8s

.PHONY: generate_crds
generate_crds:
	operator-sdk generate crds

.PHONY: template_deployment
template_deployment:
	sed "s|REPLACE_IMAGE|$(APP_IMAGE)|g" deploy-templates/deployment.yaml > deploy/deployment.yaml

.PHONY: build
build: generate_code
	operator-sdk build $(APP_IMAGE)

.PHONY: apply
apply: build template_deployment generate_crds
	kind load docker-image $(APP_IMAGE) --name $(CLUSTER_NAME)
	-kubectl delete -f ./deploy/deployment.yaml
	kubectl apply --recursive -f ./deploy

.PHONY: integration_test
integration_test: build template_deployment generate_crds
	-make kind_create
	kind load docker-image $(APP_IMAGE) --name $(CLUSTER_NAME)
	-kubectl create namespace $(TESTING_NAMESPACE)
	kubectl apply --recursive -f ./deploy -n $(TESTING_NAMESPACE)
	kubectl delete -- recursive -f ./deploy -n $(TESTING_NAMESPACE)
	kubectl delete namespace $(TESTING_NAMESPACE)

.PHONY: unit_test
unit_test: generate_code
	go test ./...

.PHONY: run
run: generate_code
	OPERATOR_NAME=$(APP_NAME) operator-sdk run --local
