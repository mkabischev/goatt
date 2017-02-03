package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mguzelevich/goatt"
)

func main() {
	// debug := flag.Bool("debug", false, "debug mode")
	dryRun := flag.Bool("dry-run", false, "dry-run mode")

	yamlFile := flag.String("yaml", "", "yaml scenario")

	publishSubject := flag.String("publish", "", "nats publish subject")
	requestSubject := flag.String("request", "", "nats request subject")
	subscription := flag.String("subscription", "", "nats subscription subject")
	queue := flag.String("queue", "", "nats queue")

	flag.Parse()

	scenario := goatt.NewScenario()

	switch {
	// case *debug:
	// 	fmt.Fprintf(os.Stderr, "debug mode")
	case *yamlFile != "":
		fmt.Fprintf(os.Stderr, "scenario mode\n")
		contents, _ := loadFile(*yamlFile)
		scenario.Load(contents)
	case *publishSubject != "":
		contents, _ := loadFile(flag.Args()[0])
		s := goatt.ScenarioStep{
			Subject: *publishSubject,
			Type:    "publish",
			Msg:     string(contents),
		}
		scenario.Steps = append(scenario.Steps, s)
	case *requestSubject != "":
		contents, _ := loadFile(flag.Args()[0])
		s := goatt.ScenarioStep{
			Subject: *requestSubject,
			Type:    "request",
			Msg:     string(contents),
		}
		scenario.Steps = append(scenario.Steps, s)
	case *subscription != "":
		s := goatt.ScenarioStep{
			Subject: *subscription,
			Type:    "subscribe",
			Msg:     *queue,
		}
		scenario.Steps = append(scenario.Steps, s)
	default:
		fmt.Fprintf(os.Stderr, "unknown mode")
	}
	scenario.Play(*dryRun)
}
