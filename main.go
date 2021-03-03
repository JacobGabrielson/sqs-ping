package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
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

func urlFor(ctx context.Context, client *sqs.Client, id string) (queueURL *url.URL, err error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("empty identifier")
	}
	if queueURL, err = url.Parse(id); err != nil {
		return
	}
	var out *sqs.GetQueueUrlOutput
	if out, err = client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(id),
	}); err != nil {
		return
	}
	rawurl := out.QueueUrl
	return url.Parse(*rawurl)
}

type localStatus struct {
	Hostname  string
	Timestamp string
}

func main() {
	var in io.Reader
	var err error

	fileName := flag.String("f", "", "file to send")
	count := flag.Int("c", 1, "send the message this many times")
	interval := flag.Duration("i", time.Millisecond*200, "seconds to wait between sends")
	flag.Parse()

	switch *fileName {
	case "":
		hostname, err := os.Hostname()
		if err != nil {
			hostname = fmt.Sprintf("unknown (%s)", err.Error())
		}
		bs, err := json.MarshalIndent(localStatus{
			Hostname:  hostname,
			Timestamp: time.Now().Format(time.RFC1123Z),
		}, "", "  ")
		if err != nil {
			log.Fatalf("creating info %v", err)
		}
		in = bytes.NewReader(bs)
	case "-":
		in = os.Stdin
	default:
		inFile, err := os.Open(*fileName)
		if err != nil {
			log.Fatalf("unable to open file '%s': %v", *fileName, err)
		}
		defer inFile.Close()
		in = inFile
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load AWS configuration, %v", err)
	}
	sqsClient := sqs.NewFromConfig(cfg)
	queueId := flag.Arg(0)
	queueURL, err := urlFor(context.TODO(), sqsClient, queueId)
	if err != nil {
		log.Fatalf("unable to find queue URL for '%s': %v", queueId, err)
	}
	for i := 0; i < *count; i++ {
		if err = send(context.TODO(), sqsClient, queueURL, in); err != nil {
			log.Fatalf("unable to send message: %v", err)
		}
		if i > 0 && i < (*count-1) {
			time.Sleep(*interval)
		}
	}
}
