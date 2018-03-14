package main

import (
	"log"
	"os"
	"strings"

	"github.com/lsegal/dinit"
)

type client struct {
	log *log.Logger
	svc lister
}

func (c *client) PrintPeople() {
	people := c.svc.ListPeople()
	c.log.Println("People:", strings.Join(people, ", "))
}

type lister interface {
	ListPeople() []string
}

type service struct {
	log *log.Logger
}

func (s *service) ListPeople() []string {
	s.log.Println("Client asked for a list of people")
	return []string{"Sarah", "Bob", "André"}
}

func newClient(l *log.Logger, svc lister) *client {
	l.Println("Initializing client")
	return &client{log: l, svc: svc}
}

func newService(l *log.Logger) *service {
	l.Println("Initializing service")
	return &service{log: l}
}

func main() {
	l := log.New(os.Stdout, "", log.Lshortfile)
	useClient := func(c *client) { c.PrintPeople() }
	dinit.Init(newClient, useClient, newService, l)

	// Output:
	// main.go:40: Initializing service
	// main.go:35: Initializing client
	// main.go:30: Client asked for a list of people
	// main.go:18: People: Sarah, Bob, André
}
