package main

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PodVerifier struct {
	clientset *kubernetes.Clientset
	namespace string
}

func (v *PodVerifier) exists(name string, label Label) (bool, error) {
	pod, err := v.clientset.CoreV1().
		Pods(v.namespace).
		Get(context.TODO(), name, metav1.GetOptions{})

	if err != nil {
		return false, err
	}

	return podHasLabel(pod, label), nil
}

func podHasLabel(pod *corev1.Pod, l Label) bool {
	labels := pod.Labels
	return labels[l.Name] == l.Value
}
