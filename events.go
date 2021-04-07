package main

import (
	"context"
	"log"

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

func watchEvents(clientset *kubernetes.Clientset, namespace string) (<-chan watch.Event, error) {
	resultChan, err := clientset.CoreV1().
		Events(namespace).
		Watch(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return nil, err
	}
	return resultChan.ResultChan(), nil
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

func filter(clientset *kubernetes.Clientset, namespace string, label Label, callback func(*EventLog)) error {
	eventsChan, err := watchEvents(clientset, namespace)
	if err != nil {
		return err
	}

	for {
		if e, ok := <-eventsChan; ok {
			event := toEvent(e.Object)

			finder := selectFinder(clientset, namespace, event.InvolvedObject.Kind)

			if finder == nil {
				log.Printf("Resource Type '%s' Not Supported Yet\n", event.InvolvedObject.Kind)
				continue
			}

			exists, err := finder.exists(event.InvolvedObject.Name, label)
			if err != nil {
				log.Println(err)
				continue
			}

			eventLog := newEventLog(event)
			if exists {
				log.Println(eventLog)
				callback(eventLog)
			}
		} else {
			log.Println("Stop watching")
			break
		}
	}
	return nil
}
