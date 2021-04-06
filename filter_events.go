package main

import (
	"context"
	"fmt"

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

func filterEventsByPodName(clientset *kubernetes.Clientset, podName string, namespace string) ([]*EventLog, error) {
	fieldSelector := fmt.Sprintf("involvedObject.kind=Pod,involvedObject.name=%s", podName)
	events, err := clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{FieldSelector: fieldSelector})
	return toEventLogList(events), err
}

func watchEventsByPodName(clientset *kubernetes.Clientset, namespace string) (watch.Interface, error) {
	return clientset.CoreV1().
		Events(namespace).
		Watch(context.TODO(), metav1.ListOptions{})
}

func toEventLogList(eventList *corev1.EventList) []*EventLog {
	var eventLogList []*EventLog
	for _, item := range eventList.Items {
		eventLogList = append(eventLogList, newEventLog(&item))
	}
	return eventLogList
}

func newEventLog(event *corev1.Event) *EventLog {
	return &EventLog{
		Type:    event.Type,
		Reason:  event.Reason,
		Object:  event.InvolvedObject.Kind + "/" + event.InvolvedObject.Name,
		Message: event.Message,
	}
}
