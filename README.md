# K8s Operators
K8sOperators is a playground based on kind (Kubernetes in Docker) in which
to experiment with the Operator SDK. It also serves as an example of how
an operator-based project might be structured to support end-to-end
integration testing.

## Components used in setting up this project
```
go 1.13.5
operator-sdk 0.16.0
kind 0.7.0 (0.7.0+ is required for disk access)
kubectl 1.17.3
```

## How operators work
### Predicates

## Operators in this project
### ManagedNamespace


## Testing


## Commands
```
Create kind cluster:
make kind_create

Destroy kind cluster:
make kind_destroy

Build Go binary and package into Docker image:
make build

Run unit tests:
make unit_test

Run integration tests:
make integration_test

Run operator locally:
make run
```
