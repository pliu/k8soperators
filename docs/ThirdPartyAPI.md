# Third Party API
ThirdPartyAPI is meant to be an example of how to use third-party APIs in
our projects (we use SealedSecrets in this example).

Firstly, the SealedSecret CRD is installed along with the other K8sOperator
resources. Examples of how to import the SealedSecret API to create, get,
and delete SealedSecrets can be found in the ThirdPartyAPI server
controller.

## User guide
The ThirdPartyAPI server controller uses the `/tpa` route prefix. The
ThirdPartyAPI server controller has three endpoints, `/create`, `/get`, and
`/delete` for creating, getting, and deleting SealedSecrets. These endpoints
process POST requests with a JSON payload containing a single `name` field.

```
curl -H "Content-Type: application/json" localhost:8080/tpa/create -d '{"name":"<resource name>"}'
curl localhost:8080/tpa/get/<resource name>
curl -X DELETE localhost:8080/tpa/delete
```

If the operator is running in the cluster instead of locally, use port
8181\.
