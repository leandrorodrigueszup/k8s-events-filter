package main

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func findPod(clientset *kubernetes.Clientset, namespace string, name string) (*corev1.Pod, error) {
	return clientset.CoreV1().
		Pods(namespace).
		Get(context.TODO(), name, metav1.GetOptions{})
}
