package constants

import (
	"os"
)

const (
	K8sOperatorsLabelKey       = "k8soperators.pliu.github.com"
	ManagedNamespaceLabelValue = "managednamespace"
	ManagedNamespaceName       = "anchor"
	ThirdPartyAPILabelValue    = "thirdpartyapi"
)

var (
	OperatorNamespace = "default"
)

func init() {
	ns, found := os.LookupEnv("OPERATOR_NAMESPACE")
	if found {
		OperatorNamespace = ns
	}
}
