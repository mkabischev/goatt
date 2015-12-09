package main

import ()

type Client interface {
	Init(string, string)
	Publish(*Context, ScenarioStep, bool)
	Request(*Context, ScenarioStep, bool)
	Subscribe(*Context, ScenarioStep, bool)
}
