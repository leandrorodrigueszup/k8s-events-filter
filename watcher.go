package main

import (
	"context"
	"encoding/json"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

type EventDTO struct {
	Type    string
	Reason  string
	Object  string
	Message string
}

type Kind struct {
	name     string
	verifier ResourceVerifier
}

type Watcher struct {
	clientset *kubernetes.Clientset
	namespace string
	label     Label
	observed  map[string]ResourceVerifier
}

func NewWatcher(clientset *kubernetes.Clientset, namespace string, label Label) *Watcher {
	return &Watcher{
		clientset: clientset,
		namespace: namespace,
		label:     label,
		observed:  make(map[string]ResourceVerifier),
	}
}

func (w *Watcher) For(kind *Kind) *Watcher {
	w.observed[kind.name] = kind.verifier
	return w
}

func (w *Watcher) Start(cb func(*EventDTO)) error {
	eventsChan, err := w.clientset.CoreV1().
		Events(w.namespace).
		Watch(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return err
	}

	for {
		if e, ok := <-eventsChan.ResultChan(); ok {
			event := toEvent(e.Object)

			if verifier, ok := w.observed[event.InvolvedObject.Kind]; ok {
				// Here could have a cache for events from the same resource multiple times
				exists, err := verifier.exists(event.InvolvedObject.Name, w.label)
				if err != nil {
					log.Println(err)
					continue
				}
				if exists {
					cb(newEventDTO(event))
				}
			} else {
				log.Printf("Resource Type '%s' Not Supported Yet\n", event.InvolvedObject.Kind)
				continue
			}
		} else {
			log.Println("Stop watching")
			break
		}
	}
	return nil
}

func newEventDTO(event *corev1.Event) *EventDTO {
	return &EventDTO{
		Type:    event.Type,
		Reason:  event.Reason,
		Object:  formatObjectDescription(event.InvolvedObject),
		Message: event.Message,
	}
}

func formatObjectDescription(involvedObject corev1.ObjectReference) string {
	return involvedObject.Kind + "/" + involvedObject.Name
}

func toEvent(object runtime.Object) *corev1.Event {
	bytes, _ := json.Marshal(object)
	var event corev1.Event
	json.Unmarshal(bytes, &event)
	return &event
}
