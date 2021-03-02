package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/JacobGabrielson/sqs-ping/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	_ "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func send(ctx context.Context, client *sqs.Client, queueURL fmt.Stringer, reader io.Reader) error {
	bs, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	_, err = client.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody:  aws.String(string(bs)),
		QueueUrl:     aws.String(queueURL.String()),
		DelaySeconds: 0,
	})
	return err
}

// Features:
// - timeout
// - repeat
// - delay between timeouts
// - parallel
// - send file contents
// - send multiple file contents
// - generate fake data (of size)
// - support FIFO queues

func main() {
	//flag.String("region", "", "AWS region (defaults to local region)") // TODO: impl
	flag.Parse()

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load AWS configuration, %v", err)
	}
	sqsClient := sqs.NewFromConfig(cfg)
	queueId := flag.Arg(0)
	queueURL, err := url.For(context.TODO(), sqsClient, queueId)
	if err != nil {
		log.Fatalf("unable to find queue URL for '%s': %v", queueId, err)
	}
	if err = send(context.TODO(), sqsClient, queueURL, os.Stdin); err != nil {
		log.Fatalf("unable to send message: %v", err)
	}
}
