package controllers

import (
	"context"
	"fmt"
	"io"
	"k8soperators/pkg/constants"
	"k8soperators/pkg/server/utils"
	"net/http"
	"strings"

	sealedsecretsv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	sealedsecretsclientset "github.com/bitnami-labs/sealed-secrets/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	thirdPartyAPIController = Controller{
		Name: "ThirdPartyAPIController",
		Path: "/tpa",
		Mux:  http.NewServeMux(),
	}
	tpaLabels = labels.Set(map[string]string{
		constants.K8sOperatorsLabelKey: constants.ThirdPartyAPILabelValue,
	})
	sealedSecretsClientset *sealedsecretsclientset.Clientset = nil
)

type ThirdPartyAPIBody struct {
	Name string
}

var tpaLog = logf.Log.WithName(thirdPartyAPIController.Name)

func createSealedSecret(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only accepting POST requests", http.StatusMethodNotAllowed)
		return
	}

	var body ThirdPartyAPIBody
	if err := utils.GetJson(w, r, &body); err != nil {
		tpaLog.Info(err.Error())
		return
	}

	sealedsecret := &sealedsecretsv1alpha1.SealedSecret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      body.Name,
			Namespace: constants.OperatorNamespace,
			Labels:    tpaLabels,
		},
		Spec: sealedsecretsv1alpha1.SealedSecretSpec{
			EncryptedData: map[string]string{
				"test": "abc",
			},
		},
	}

	if err := k8sClient.Create(context.TODO(), sealedsecret); err != nil {
		tpaLog.Info(fmt.Sprintf("Failed to create SealedSecret %s", sealedsecret.Name))
		http.Error(w, fmt.Sprintf("Failed to create SealedSecret: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	tpaLog.Info(fmt.Sprintf("Created SealedSecret %s", sealedsecret.Name))

	io.WriteString(w, fmt.Sprintf("SealedSecret created: %s", sealedsecret.Name))
}

func getSealedSecret(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only accepting GET requests", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/get/")

	if sealedSecretsClientset == nil {
		sealedSecretsClientset = sealedsecretsclientset.NewForConfigOrDie(k8sConfig)
		tpaLog.Info("Created SealedSecretsClientset")
	}
	sealedsecret, err := sealedSecretsClientset.BitnamiV1alpha1().SealedSecrets(constants.OperatorNamespace).Get(name, metav1.GetOptions{})
	if err != nil {
		tpaLog.Info(fmt.Sprintf("Failed to get SealedSecret %s", name))
		http.Error(w, fmt.Sprintf("Failed to get SealedSecret: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	tpaLog.Info(fmt.Sprintf("Found SealedSecret %v", sealedsecret))

	io.WriteString(w, fmt.Sprintf("SealedSecret found: %v", sealedsecret))
}

func deleteSealedSecret(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Only accepting DELETE requests", http.StatusMethodNotAllowed)
		return
	}

	sealedsecret := &sealedsecretsv1alpha1.SealedSecret{}
	deleteAllOpts := &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			Namespace:     constants.OperatorNamespace,
			LabelSelector: tpaLabels.AsSelector(),
		},
	}
	if err := k8sClient.DeleteAllOf(context.TODO(), sealedsecret, deleteAllOpts); err != nil {
		tpaLog.Info("Failed to delete SealedSecrets")
		http.Error(w, "Failed to delete SealedSecret", http.StatusInternalServerError)
		return
	}

	tpaLog.Info("Deleted SealedSecrets")

	io.WriteString(w, "SealedSecrets deleted")
}

func init() {
	thirdPartyAPIController.Mux.HandleFunc("/create", createSealedSecret)
	thirdPartyAPIController.Mux.HandleFunc("/get/", getSealedSecret)
	thirdPartyAPIController.Mux.HandleFunc("/delete", deleteSealedSecret)

	registerController(&thirdPartyAPIController)
}
