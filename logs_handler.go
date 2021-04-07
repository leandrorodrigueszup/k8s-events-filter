package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

func logsHandler(clientset *kubernetes.Clientset) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c.Response().WriteHeader(http.StatusOK)
		enc := json.NewEncoder(c.Response())

		label := extractLabel(c.QueryParam("label"))

		namespace := getNamespaceOr(c, defaultNamespace)
		watch, err := watchEventsByPodName(clientset, namespace)
		if err != nil {
			return err
		}

		for {
			if e, ok := <-watch.ResultChan(); ok {
				event := toEvent(e.Object)

				finder := selectFinder(clientset, namespace, event.InvolvedObject.Kind)

				if finder == nil {
					fmt.Printf("Resource Type '%s' Not Supported Yet\n", event.InvolvedObject.Kind)
					continue
				}

				exists, err := finder.exists(event.InvolvedObject.Name, label)
				if err != nil {
					fmt.Println(err)
					continue
				}

				eventLog := toEventLog(e.Object)
				if !exists {
					fmt.Printf("DISCART: %v\n", eventLog)
				} else {
					fmt.Println(eventLog)

					if err := enc.Encode(eventLog); err != nil {
						fmt.Println(err)
						continue
					}
				}
				c.Response().Flush()
			} else {
				fmt.Println("Stop watching")
				break
			}
		}
		return nil
	}
}

func extractLabel(l string) Label {
	keyValue := strings.Split(l, "=")
	return Label{keyValue[0], keyValue[1]}
}

func getNamespaceOr(c echo.Context, defaultNamespace string) string {
	namespace := c.QueryParam("namespace")
	if namespace == "" {
		return defaultNamespace
	}
	return namespace
}

func selectFinder(clientset *kubernetes.Clientset, namespace string, kind string) FindResource {
	switch kind {
	case "Pod":
		return &PodFinder{clientset, namespace}
	case "Service":
		return &ServiceFinder{clientset, namespace}
	case "Deployment":
		return &DeploymentFinder{clientset, namespace}
	default:
		return nil
	}
}

func toEvent(eventObject runtime.Object) *corev1.Event {
	bytes, _ := json.Marshal(eventObject)
	var event corev1.Event
	json.Unmarshal(bytes, &event)
	return &event
}

func toEventLog(eventObject runtime.Object) *EventLog {
	return newEventLog(toEvent(eventObject))
}
