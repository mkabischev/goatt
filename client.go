package goatt

import ()

type Client interface {
	Init(string, string)
	Publish(*Context, ScenarioStep, bool)
	Request(*Context, ScenarioStep, bool)
	Subscribe(*Context, ScenarioStep, bool)
}

type Clients struct {
	clients map[string]Client
}