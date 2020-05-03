# ManagedNamespace
As its name might suggest, ManagedNamespace is an operator for managing
namespaces. Specifically, it is designed to allow users to self-serve and
create and delete ephemeral namespaces while maintaining security across the
cluster's other namespaces (e.g. the ones in which applications are
running).

It does this by exposing a REST API that allows users to request an
ephemeral namespace. It then creates a namespace, a role binding
allowing only the requesting user permissions within the namespace, and a
ManagedNamespace object in the namespace. Namespaces created this way are
labeled with the K8sOperatorsLabelKey and ManagedNamespaceLabelValue
(defined in constants). The ManagedNamespace controller watches for
ManagedNamespace deletions and reconciles the cluster state by deleting any
such labeled namespaces in which a ManagedNamespace object does not exist.
As only the user who requested the ephemeral namespace (and cluster admins)
have the permissions required to delete the ManagedNamespace object within
the ephemeral namespace, this provides a mechanism for the user to request
deletion of their own ephemeral namespaces when they are done with them.

## User guide
The ManagedNamespace server controller uses the `/managednamespace` route
prefix. The ManagedNamespace server controller currently only has one
endpoint, `/create`, for creating managed namespaces. This endpoint
processes POST requests with a JSON payload containing a single `user`
field.
```
curl -H "Content-Type: application/json" localhost:8080/managednamespace/create -d '{"user":"<username>"}'
```
To delete the managed namespace, simply delete the ManagedNamespace object
in the namespace.
```
kubectl delete managednamespace --all -n <namespace>
```
