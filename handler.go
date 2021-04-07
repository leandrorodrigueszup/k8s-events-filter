package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

func eventsHandler(clientset *kubernetes.Clientset) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c.Response().WriteHeader(http.StatusOK)
		enc := json.NewEncoder(c.Response())

		label := getLabel(c.QueryParam("label"))

		namespace := getNamespaceOr(c, defaultNamespace)

		log.SetPrefix(label.Name + "[" + label.Value + "] - ")
		log.SetFlags(0)

		filter(clientset, namespace, label, func(eventLog *EventLog) {
			if err := enc.Encode(eventLog); err != nil {
				log.Println(err)
			}
			c.Response().Flush()
		})
		return nil
	}
}

func getLabel(l string) Label {
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
