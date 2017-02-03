package goatt

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type JsonMap map[string]interface{}

type ScenarioStep struct {
	Subject  string "target"
	Type     string "type"
	Msg      string "body"
	Protocol string "protocol"
}

type yamlScenario struct {
	Common    map[string]interface{} "common"
	Constants map[string]interface{} "constants"
	Steps     []ScenarioStep         ",flow"
}

var ClientNATS Client
var ClientSQS Client

func (ys *yamlScenario) Load(contents []byte) error {
	if err := yaml.Unmarshal(contents, ys); err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse YAML scenario file: %s\n", err)
		return err
	}
	return nil
}

func (ys *yamlScenario) natsServer() string {
	if nats, ok := ys.Common["server_nats"]; ok && nats.(string) != "" {
		return nats.(string)
	}
	if ys.Common["protocol"] == "nats" {
		return ys.Common["server"].(string)
	}
	return ""
}

func (ys *yamlScenario) sqsServer() string {
	if sqs, ok := ys.Common["server_sqs"]; ok && sqs.(string) != "" {
		return sqs.(string)
	}
	if ys.Common["protocol"] == "sqs" {
		return ys.Common["server"].(string)
	}
	return ""
}

func (ys *yamlScenario) Play(dryRun bool) error {
	if nats := ys.natsServer(); nats != "" {
		ClientNATS.Init(nats, ys.Common["service"].(string))
	}
	if sqs := ys.sqsServer(); sqs != "" {
		ClientSQS.Init(sqs, ys.Common["service"].(string))
	}

	timeout := 1 * time.Microsecond
	tms := ys.Common["timeout"]
	if tms != nil {
		if value, err := time.ParseDuration(tms.(string)); err == nil {
			timeout = value
		}
	}

	var protocol string
	if proto := ys.Common["protocol"]; proto != nil {
		protocol = proto.(string)
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

		var client Client
		currentProtocol := protocol
		if step.Protocol != "" {
			currentProtocol = step.Protocol
		}

		switch currentProtocol {
		case "nats":
			client = ClientNATS
		case "sqs":
			client = ClientSQS
		default:
			panic("invalid protocol")
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

func NewScenario() *yamlScenario {
	scenario := new(yamlScenario)
	ClientNATS = new(NatsClient)
	ClientSQS = new(SQSClient)
	return scenario
}
