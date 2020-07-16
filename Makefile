CLUSTER_NAME=k8soperators
KIND_IMAGE=kindest/node:v1.18.2
APP_NAME=k8soperators
VERSION=$$(grep -o '".*"' version/version.go | sed 's/"//g')
APP_IMAGE=$(APP_NAME):$(VERSION)

.PHONY: kind_create
kind_create:
	kind create cluster --config=kind/config.yaml --name $(CLUSTER_NAME) --image $(KIND_IMAGE)

.PHONY: kind_destroy
kind_destroy:
	kind delete cluster --name $(CLUSTER_NAME)

.PHONY: generate_code
generate_code:
	operator-sdk generate crds
	operator-sdk generate k8s

.PHONY: template_deployment
template_deployment:
	sed "s|REPLACE_IMAGE|$(APP_IMAGE)|g" deploy-templates/deployment.yaml > deploy/deployment.yaml

.PHONY: build
build: generate_code
	operator-sdk build $(APP_IMAGE)

.PHONY: apply
apply: build template_deployment
	kind load docker-image $(APP_IMAGE) --name $(CLUSTER_NAME)
	-kubectl delete -f deploy/deployment.yaml
	kubectl apply --recursive -f deploy/

.PHONY: delete
delete:
	kubectl delete --recursive -f deploy/

.PHONY: integration_tests
integration_tests:
	-make apply
	sleep 20
	go test -count=1 ./integration-tests/...

.PHONY: unit_tests
unit_tests: generate_code
	go test -count=1 ./pkg/...

.PHONY: run
run: generate_code
	-kubectl apply --recursive -f deploy/crds
	OPERATOR_NAME=$(APP_NAME) operator-sdk run --local --namespace=''
