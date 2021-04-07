package main

type Label struct {
	Name  string
	Value string
}

type ResourceVerifier interface {
	exists(name string, label Label) (bool, error)
}