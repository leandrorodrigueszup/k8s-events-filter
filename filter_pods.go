package main

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func filterPodsByLabel(clientset *kubernetes.Clientset, labels map[string]string, namespace string) ([]string, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), prepareListOptions(labels))
	if err != nil {
		return nil, err
	}

	return mapPod(pods.Items, func(p corev1.Pod) string {
		return p.Name
	}), nil
}

func prepareListOptions(labels map[string]string) metav1.ListOptions {
	if len(labels) == 0 {
		return metav1.ListOptions{}
	}
	labelSelector := formatLabelQuery(labels)
	return metav1.ListOptions{LabelSelector: labelSelector}
}

func formatLabelQuery(labels map[string]string) string {
	var list []string
	for k, v := range labels {
		list = append(list, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(list, ",")
}

func mapPod(items []corev1.Pod, mapping func(corev1.Pod) string) []string {
	var result = make([]string, len((items)))
	for i, item := range items {
		result[i] = mapping(item)
	}
	return result
}
