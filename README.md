# K8s Operators
K8sOperators is a playground based on kind (Kubernetes in Docker) in which
to experiment with the Operator SDK. It also serves as an example of how an
operator-based project might be structured to support end-to-end
integration testing.

This project assumes basic familiarity with Kubernetes. If you are new to
Kubernetes, it would be helpful to read about the high-level model of how
Kubernetes works (e.g. how it models the actual state of a cluster, both
the hardware and the applications that are running within it, and how it
constantly reconciles its observed state with some desired state). There
is also a sister project to help familiarize oneself with Kubernetes:
https://github.com/pliu/k8splayground.

This is a single operator that contains numerous controllers. This was
mainly because of laziness - test multiple ideas for controllers in a
single project without having to duplicate the scaffolding (e.g., kind,
Makefile, existing examples). As the core logic is captured within the
controllers, which are mostly independent of each other, if, in the future,
the functionality of a given controller is deemed worth deploying as its
own operator, it would be fairly simple to break that functionality out.

One downside of using a single operator is that all controllers within it
will share the same scope (i.e. cluster or namespace) - K8sOperators is
cluster-scoped. As being cluster-scoped requires handling a superset of the
cases handled by namespace-scoped controllers, splitting a controller from
K8sOperators out into its own operator that is namespace-scoped should be
simple.

## Components used in setting up this project
```
go 1.13.5, 1.17.1
operator-sdk 0.16.0
kind 0.7.0 (0.7.0+ is required for disk access)
kubectl 1.17.3
```

## How operators work
#### Manager
The manager is the core component of the operator framework. It sets up the
scaffolding for common controller activities (e.g., providing the
Kubernetes client, exposing metrics through an endpoint that can be scraped
by Prometheus).

K8sOperators makes use of the metrics server that comes with the manager,
registering controller-specific metrics with the registry the metrics
server serves.

To enable access to the server when it is deployed in Kubernetes, when
creating the cluster, we port-forward localhost port 8484 on the local
machine to port 30001 on one of the Docker worker "hosts". A NodePort
service then forwards all host traffic on port 30001 to the server on port
8383\. Thus to access the metrics server:
```
Operator is running locally:
curl localhost:8383/metrics

Operator is running in the cluster:
curl localhost:8484/metrics
```

The manager has a SyncPeriod setting that sets the interval at which all
watched objects in the cache are automatically reconciled. Using a data
processing analogy, the state is normally updated in a streaming manner -
with each received event. Over time, however, since events can be lost, the
materialized state can diverge from the true state. To resolve this, one
can run batch jobs at some specified interval to reconcile the true,
point-in-time state, after which subsequent events mutate the newly
converged state in a streaming manner again. The SyncPeriod is essentially
the interval at which the batch job would be run.

We have essentially disabled this functionality, however, by setting the
period to 10 years. The reason is that, for each controller, the controller
either triggers a global reconciliation on each event or has a background
process that performs global reconciliation. Thus, this functionality is
not required to maintain state convergence.

#### Controllers
Controllers comprise the core logic of any operator. They work by
subscribing to events (e.g., creation, update, or deletion of specific
Kubernetes resources) and triggering a reconciliation action whenever
qualifying events occur.

More specifically, they use Source objects to subscribe to events of
specific resource types, Handler objects to build the reconciliation
Request object, and a Predicate to filter for qualifying events.
Reconciliation requests are queued and handled by the controller's user-
defined Reconcile method which takes the appropriate actions to move the
cluster to the desired state (it can take other actions as well, e.g.
sending notifications). The Reconcile method can also choose to requeue the
request to be reprocessed later. It is recommended that the Reconcile
method be idempotent, thus conforming to Kubernetes' philosophy of
declarative management (if a controller was imperative, if its managed
state deviated from the desired state, manual intervention would be
required to resolve the inconsistency).

Controllers must register themselves with the manager, similar to
callbacks.

#### Custom resources
Kubernetes allows users to define their own resources using Custom Resource
Definitions (CRDs). The operator-sdk generates CRD manifests based on a Go
type definitions. The operator-sdk also autogenerates code that, similar to
an ORM, maps between Go objects and Kubernetes objects and vice-versa.
Similar to controllers, custom resources must be registered with the
manager.

#### Server
Although not part of the Operator SDK, K8sOperators comes with a server to
allow for external interactions with the controllers in the project (e.g.
the ManagedNamespace operator controller manages deletion of managed
namespaces while the ManagedNamespace server controller manages user-
requested managed namespace creation, together providing the complete
ManagedNamespace functionality). Just as it is necessary to write a
separate operator controller for each operator functionality, a separate
server controller is recommended for each server functionality.

The single server running on port 8080 is shared amongst the controllers
using different path prefixes. To enable access to the server when it is
deployed in Kubernetes, when creating the cluster, we port-forward
localhost port 8181 on the local machine to port 30000 on one of the Docker
worker "hosts". A NodePort service then forwards all host traffic on port
30000 to the server on port 8080.

#### Background
Also not a part of the Operator SDK, K8sOperators comes with a framework for
running global reconciliations in the background, including initial global
reconciliations at operator start. The purpose of this is to reconcile any
changes that controllers missed in between global reconciliations, including
any that occurred while the operator was unavailable.

## Controllers in this project
[ManagedNamespace](docs/ManagedNamespace.md)
[Third Party API](docs/ThirdPartyAPI.md)

## Testing
There are numerous methods in K8sOperators to hasten the feedback loop of
testing to accelerate development. These include running the operator
locally against a cluster, unit tests, integration tests, and applying the
operator along with other resources it depends on (e.g., ServiceAccount,
Role, RoleBinding) to a cluster. The cluster used is the one specified by
the kubeconfig found at ~/.kube/config.

When initially developing a controller, running the operator locally allows
for rapid iteration (this runs the current code and displays the log
messages). While running locally, any endpoints exposed by the operator
(e.g. the metrics endpoint) are accessible on localhost. As the
functionality sets, it should be codified in unit and integration tests that
can be run on subsequent code changes to ensure that the desired behavior
was not altered. Finally, the operator can be applied to the cluster as it
might be in a live setting.

## Commands
```
Create kind cluster:
make kind_create

Destroy kind cluster:
make kind_destroy

Run operator locally against the cluster:
make run

Run unit tests:
make unit_tests

Run integration tests:
make integration_tests

Apply operator to the cluster:
make apply

Delete operator from the cluster:
make delete

Build Go binary and package into Docker image:
make build
```
