package main

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PodFinder struct {
	clientset *kubernetes.Clientset
	namespace string
}

func (f *PodFinder) exists(name string, label Label) (bool, error) {
	pod, err := f.clientset.CoreV1().
		Pods(f.namespace).
		Get(context.TODO(), name, metav1.GetOptions{})

	if err != nil {
		return false, err
	}

	return podHasLabel(pod, label), nil
}

func PodHasLabel(pod *corev1.Pod, l Label) bool {
	labels := pod.Labels
	return labels[l.Name] == l.Value
}
