package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"k8s.io/client-go/kubernetes"
)

func eventsHandler(clientset *kubernetes.Clientset) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c.Response().WriteHeader(http.StatusOK)
		enc := json.NewEncoder(c.Response())

		label := getLabel(c)
		namespace := getNamespaceOr(c, defaultNamespace)

		log.SetPrefix(label.Name + "[" + label.Value + "] - ")
		log.SetFlags(0)

		watcher := NewWatcher(clientset, namespace, label)

		watcher.
			For(&Kind{"Pod", &PodVerifier{clientset, namespace}}).
			For(&Kind{"Service", &ServiceVerifier{clientset, namespace}}).
			For(&Kind{"Deployment", &DeploymentVerifier{clientset, namespace}}).
			Start(func(e *EventDTO) {
				if err := enc.Encode(e); err != nil {
					log.Println(err)
				}
				c.Response().Flush()
			})
		return nil
	}
}

func getLabel(c echo.Context) Label {
	l := c.QueryParam("label")
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
