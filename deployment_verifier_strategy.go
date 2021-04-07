package main

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type DeploymentVerifier struct {
	clientset *kubernetes.Clientset
	namespace string
}

func (v *DeploymentVerifier) exists(name string, label Label) (bool, error) {
	deployment, err := v.clientset.AppsV1().
		Deployments(v.namespace).
		Get(context.TODO(), name, metav1.GetOptions{})

	if err != nil {
		return false, err
	}

	return deploymentHasLabel(deployment, label), nil
}

func deploymentHasLabel(deployment *appsv1.Deployment, l Label) bool {
	labels := deployment.Labels
	return labels[l.Name] == l.Value
}
