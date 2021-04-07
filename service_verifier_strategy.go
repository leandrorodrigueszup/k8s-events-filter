package main

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ServiceVerifier struct {
	clientset *kubernetes.Clientset
	namespace string
}

func (v *ServiceVerifier) exists(name string, label Label) (bool, error) {
	service, err := v.clientset.CoreV1().
		Services(v.namespace).
		Get(context.TODO(), name, metav1.GetOptions{})

	if err != nil {
		return false, err
	}

	return serviceHasLabel(service, label), nil
}

func serviceHasLabel(service *corev1.Service, l Label) bool {
	labels := service.Labels
	return labels[l.Name] == l.Value
}
