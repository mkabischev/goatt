package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type JsonMap map[string]interface{}

type ScenarioStep struct {
	Subject string "target"
	Type    string "type"
	Msg     string "body"
}

type yamlScenario struct {
	Common    map[string]interface{} "common"
	Constants map[string]interface{} "constants"
	Steps     []ScenarioStep         ",flow"
}

func (ys *yamlScenario) load(contents []byte) error {
	if err := yaml.Unmarshal(contents, ys); err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse YAML scenario file: %s\n", err)
		return err
	}
	return nil
}

func (ys *yamlScenario) play(dryRun bool) error {
	client := MQClient
	client.Init(ys.Common["server"].(string), ys.Common["service"].(string))

	timeout := 1 * time.Microsecond
	tms := ys.Common["timeout"]
	if tms != nil {
		if value, err := time.ParseDuration(tms.(string)); err == nil {
			timeout = value
		}
	}

	ctx := InitContext(ys.Constants)

	for i, step := range ys.Steps {
		time.Sleep(timeout)
		ctx.ClearStep()
		ctx.Session["stepIdx"] = i + 1

		if subj, err := ctx.Evaluate(string(step.Subject)); err != nil {
			fmt.Fprintf(os.Stderr, "\n[%05d]: %s->%s ", ctx.Session["stepIdx"], step.Type, "<ERROR>")
			fmt.Fprintf(os.Stderr, "Could not evaluate message subject: %s\n", err)
			continue
		} else {
			step.Subject = subj
		}

		fmt.Fprintf(os.Stderr, "\n[%05d]: %s->%s ", ctx.Session["stepIdx"], step.Type, step.Subject)
		requestType := ys.Common["method"].(string)
		if step.Type != "" {
			requestType = step.Type
		}
		switch requestType {
		case "publish":
			client.Publish(ctx, step, dryRun)
		case "request":
			client.Request(ctx, step, dryRun)
		case "subscription":
			client.Subscribe(ctx, step, dryRun)
		default:
			fmt.Fprintf(os.Stderr, "unknown mode")
		}
	}
	return nil
}

func newScenario() *yamlScenario {
	scenario := new(yamlScenario)
	return scenario
}
