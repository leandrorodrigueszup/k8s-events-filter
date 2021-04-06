package main

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type EventLog struct {
	Type    string
	Reason  string
	Object  string
	Message string
}

func watchEventsByPodName(clientset *kubernetes.Clientset, namespace string) (watch.Interface, error) {
	return clientset.CoreV1().
		Events(namespace).
		Watch(context.TODO(), metav1.ListOptions{})
}

func newEventLog(event *corev1.Event) *EventLog {
	return &EventLog{
		Type:    event.Type,
		Reason:  event.Reason,
		Object:  formatObjectDescription(event.InvolvedObject),
		Message: event.Message,
	}
}

func formatObjectDescription(involvedObject corev1.ObjectReference) string {
	return involvedObject.Kind + "/" + involvedObject.Name
}