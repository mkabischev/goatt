package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SQSClient struct {
	Conn    *sqs.SQS
	Service string
}

var (
	INT64_ZERO        = int64(0)
	DELAY_MSG_SECONDS = int64(20)
	SQS_REQ_TIMEOUT   = 21 * time.Second
)

func (sqsClient *SQSClient) Init(server, service string) {
	disableSSL := true
	config := aws.Config{
		DisableSSL:  &disableSSL,
		Endpoint:    &server,
		Region:      aws.String("elasticmq"),
		Credentials: credentials.NewStaticCredentials("x", "x", ""),
	}
	s3Session := session.New(&config)
	sqsClient.Conn = sqs.New(s3Session)
	sqsClient.Service = service
}

func (sqsClient *SQSClient) Publish(ctx *Context, step ScenarioStep, dryRun bool) {
	var err error
	var req JsonMap
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

	// create queue
	queueUrl := sqsClient.createQueue(step.Subject)

	// publish msg
	meta, _ := req["meta"].(map[string]interface{})
	st, ok := meta["start_time"]
	var startTime int64
	if ok {
		startTime, err = strconv.ParseInt(st.(string), 10, 0)
		if err != nil {
			fmt.Fprintf(os.Stdout, "")
			return
		}
	}

	in := sqs.SendMessageInput{
		DelaySeconds: CalcNextDelay(&startTime),
		MessageBody:  &body,
		QueueUrl:     queueUrl,
	}

	fmt.Fprintf(os.Stderr, "send to sqs msg %q\n", in)
	_, err = sqsClient.Conn.SendMessage(&in)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
}

func (sqsClient *SQSClient) Request(ctx *Context, step ScenarioStep, dryRun bool) {
	var err error

	var req JsonMap
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

	// create queue for request
	stepQueueUrl := sqsClient.createQueue(step.Subject)

	// publish to sqs
	meta, _ := req["meta"].(map[string]interface{})
	replyQueueName, ok := meta["reply_queue"]
	if !ok {
		fmt.Fprintf(os.Stderr, "no replay_queue url in meta.")
		return
	}

	// create queue to handle response
	responseQueueUrl := sqsClient.createQueue(sqsClient.Service + "_" + replyQueueName.(string))

	// delete temp queue after all
	defer sqsClient.Conn.DeleteQueue(&sqs.DeleteQueueInput{QueueUrl: responseQueueUrl})

	smin := sqs.SendMessageInput{
		MessageBody: &body,
		QueueUrl:    stepQueueUrl,
	}
	_, err = sqsClient.Conn.SendMessage(&smin)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}

	// receive respose from sqs
	respCh := make(chan interface{})
	errCh := make(chan interface{})
	defer close(respCh)
	defer close(errCh)
	rmin := sqs.ReceiveMessageInput{WaitTimeSeconds: &DELAY_MSG_SECONDS, QueueUrl: responseQueueUrl}
	go func() {
		for {
			rmout, err := sqsClient.Conn.ReceiveMessage(&rmin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to received response from sqs queue [%s]\n", *responseQueueUrl)
				errCh <- nil
				return
			}
			switch len(rmout.Messages) {
			case 0:
				fmt.Fprintf(os.Stderr, "received 0 messages.\n")
			case 1:
				data := rmout.Messages[0].Body

				fmt.Fprintf(os.Stdout, "unpack request: body=%s\n", *data)
				var resp interface{}
				if err := json.Unmarshal([]byte(*data), &resp); err != nil {
					fmt.Fprintf(os.Stderr, "unmarshal handle response message err [%s]\n", err)
					errCh <- nil
					return
				}
				fmt.Fprintf(os.Stderr, "unmarshalled body [%s]\n", resp)

				respCh <- resp
				return
			default:
				fmt.Fprintf(os.Stderr, "received incorrect number of messages for sync sqs request to queue [%s]\n", responseQueueUrl)
				errCh <- nil
				return
			}
		}
	}()

	select {
	case resp := <-respCh:
		ctx.Result = resp
		fmt.Printf("rsp: %v\n", resp)
	case <-errCh:
	case <-time.After(SQS_REQ_TIMEOUT):
		fmt.Fprintf(os.Stderr, "timeout on waiting resp from sqs\n")
	}

	return
}

func (sqsClient *SQSClient) Subscribe(ctx *Context, step ScenarioStep, dryRun bool) {
	fmt.Fprintf(os.Stderr, "unimplemented subscribe\n")
}

//CalcNextDelay calc next sqs delay for message
func CalcNextDelay(startTime *int64) *int64 {
	if *startTime == 0 {
		return &INT64_ZERO
	}

	shift := time.Now().Unix() - *startTime
	if shift < 0 {
		// run now
		return &INT64_ZERO
	}

	// max sqs msg delay is 900sec=15min
	if shift > 900 {
		shift = 900
	}

	return &shift
}

func (sqsClient *SQSClient) createQueue(qName string) *string {
	var qUrl *string
	gquout, err := sqsClient.Conn.GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: &qName})
	if err != nil {
		cqin := sqs.CreateQueueInput{QueueName: &qName}
		cqout, err := sqsClient.Conn.CreateQueue(&cqin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error on queue create: %s\n", err.Error())
			return nil
		}
		qUrl = cqout.QueueUrl
		fmt.Fprintf(os.Stderr, "created queue %q\n", *qUrl)
	} else {
		qUrl = gquout.QueueUrl
	}

	return qUrl
}
