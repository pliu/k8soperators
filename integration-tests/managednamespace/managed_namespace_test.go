package managednamespace

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	k8soperatorsv1alpha1 "k8soperators/pkg/apis/k8soperators/v1alpha1"
	"k8soperators/pkg/constants"
	"k8soperators/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	createURL  = "http://localhost:8181/managednamespace/create"
	metricsURL = "http://localhost:8484/metrics"
	createdNamespacesMetricKey = "managednamespace_created_namespaces"
	deletedNamespacesMetricKey = "managednamespace_deleted_namespaces"
	username = "test"
)

func TestCreateDeleteManagedNamespace(t *testing.T) {
	assert := assert.New(t)
	adminClient, err := utils.GetK8sClient("")
	if err != nil {
		assert.Fail(err.Error())
	}
	impersonatedRightUserClient, err := utils.GetK8sClient(username)
	if err != nil {
		assert.Fail(err.Error())
	}
	impersonatedWrongUserClient, err := utils.GetK8sClient(username + "2")
	if err != nil {
		assert.Fail(err.Error())
	}

	assert.False(namespaceActive(adminClient, username, t))
	assert.False(managedNamespaceInNamespace(adminClient, username, t))
	data, err := utils.Get(metricsURL)
	assert.Nil(err)
	assert.True(metricIsExpected(data, createdNamespacesMetricKey, 0))
	assert.True(metricIsExpected(data, deletedNamespacesMetricKey, 0))

	assert.Nil(utils.SendPost(createURL, fmt.Sprintf("{\"user\":\"%s\"}", username)))
	time.Sleep(1 * time.Second)

	assert.True(namespaceActive(adminClient, username, t))
	managedNamespaces := &k8soperatorsv1alpha1.ManagedNamespaceList{}
	if err := impersonatedRightUserClient.List(context.TODO(), managedNamespaces, &client.ListOptions{
		Namespace: username,
	}); err != nil {
		assert.Fail(err.Error())
	}
	found := false
	var targetManagedNamespace k8soperatorsv1alpha1.ManagedNamespace
	for _, managedNamespace := range managedNamespaces.Items {
		if managedNamespace.Namespace == username {
			found = true
			targetManagedNamespace = managedNamespace
			break
		}
	}
	assert.True(found)
	data, err = utils.Get(metricsURL)
	assert.Nil(err)
	assert.True(metricIsExpected(data, createdNamespacesMetricKey, 1))
	assert.True(metricIsExpected(data, deletedNamespacesMetricKey, 0))

	assert.Error(impersonatedWrongUserClient.Delete(context.TODO(), &targetManagedNamespace))
	assert.Nil(impersonatedRightUserClient.Delete(context.TODO(), &targetManagedNamespace))
	time.Sleep(1 * time.Second)

	assert.False(namespaceActive(adminClient, username, t))
	assert.False(managedNamespaceInNamespace(adminClient, username, t))
	data, err = utils.Get(metricsURL)
	assert.Nil(err)
	assert.True(metricIsExpected(data, createdNamespacesMetricKey, 1))
	assert.True(metricIsExpected(data, deletedNamespacesMetricKey, 1))
}

func namespaceActive(c client.Client, namespaceName string, t *testing.T) bool {
	namespace := &corev1.Namespace{}
	if err := c.Get(context.TODO(), client.ObjectKey{Name: namespaceName}, namespace); err != nil {
		return false
	}
	if namespace.DeletionTimestamp == nil {
		return true
	}
	return false
}

func managedNamespaceInNamespace(c client.Client, namespaceName string, t *testing.T) bool {
	managedNamespace := &k8soperatorsv1alpha1.ManagedNamespace{}
	if err := c.Get(context.TODO(), client.ObjectKey{Name: constants.ManagedNamespaceName, Namespace: namespaceName},
		managedNamespace); err != nil {
		return false
	}
	return true
}

func metricIsExpected(data string, metricPrefix string, expected int) bool {
	metricStrings := strings.Split(data, "\n")
	metricLine := ""
	for _, metricString := range metricStrings {
		if strings.HasPrefix(metricString, metricPrefix) {
			metricLine = metricString
			break
		}
	}
	if metricLine == "" {
		return false
	}

	metric, err := strconv.Atoi(strings.Split(metricLine, " ")[1])
	if err != nil {
		return false
	}

	if metric == expected {
		return true
	}
	return false
}
