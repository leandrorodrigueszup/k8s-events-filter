package main

type Label struct {
	Name  string
	Value string
}

type FindResource interface {
	exists(name string, label Label) (bool, error)
}