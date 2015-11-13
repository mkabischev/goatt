package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/apcera/nats"
)

var timeout = 500 * time.Millisecond

type natsClient struct {
	natsConn   *nats.Conn
	natsEnConn *nats.EncodedConn
}

func (nc *natsClient) Init(server string) {
	var err error
	nc.natsConn, err = nats.Connect(server)
	if err != nil {
		panic(err)
	}
	nc.natsEnConn, err = nats.NewEncodedConn(nc.natsConn, nats.JSON_ENCODER)
	if err != nil {
		panic(err)
	}
	//defer ec.Close()
}

func (nc *natsClient) request(ctx *Context, step step, dryRun bool) {
	var err error

	var req jsonMap
	body, err := ctx.Evaluate(string(step.Msg))
	msg := []byte(body)
	if err := json.Unmarshal(msg, &req); err != nil {
		fmt.Fprintf(os.Stderr, "handle message [%s]", err)
		return
	}
	fmt.Fprintf(os.Stderr, "[%s]\n", body)
	if dryRun {
		return
	}
	response, err := nc.natsConn.Request(step.Subject, msg, timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "response message error [%s]", err)
		return
	}

	var resp interface{}
	if err := json.Unmarshal(response.Data, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "unmarshal handle response message [%s]", err)
		return
	}

	ctx.Result = resp
	if err != nil {
		fmt.Printf("rsp ERROR: [%s]\n", err)
	} else {
		fmt.Printf("rsp: [%s]\n[%s]\n", string(response.Subject), string(response.Data))

	}
}

func (nc *natsClient) publish(ctx *Context, step step, dryRun bool) {
	var err error
	var req jsonMap
	body, err := ctx.Evaluate(string(step.Msg))
	msg := []byte(body)
	if err := json.Unmarshal(msg, &req); err != nil {
		fmt.Fprintf(os.Stderr, "handle message [%s]", err)
		return
	}
	fmt.Fprintf(os.Stderr, "[%s]\n", body)
	if dryRun {
		return
	}
	if err = nc.natsEnConn.Publish(step.Subject, msg); err != nil {
		fmt.Printf("rsp ERROR: [%s]\n", err)
	}
}

func (nc *natsClient) subscribe(ctx *Context, step step, dryRun bool) {
	fmt.Fprintf(os.Stderr, "\n")
	if step.Msg != "" {
		nc.natsEnConn.QueueSubscribe(step.Subject, step.Msg, nc.handler)
	} else {
		nc.natsEnConn.Subscribe(step.Subject, nc.handler)
	}
	c := make(chan bool, 0)
	<-c
}

func (nc *natsClient) handler(msg *nats.Msg) {
	fmt.Fprintf(os.Stderr, "handle message [%s]", msg)
}
