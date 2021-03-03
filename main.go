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

func send(ctx context.Context, client *sqs.Client, queueURL fmt.Stringer, reader func() io.Reader) (int, error) {
	bs, err := ioutil.ReadAll(reader())
	if err != nil {
		return 0, err
	}
	_, err = client.SendMessage(ctx, &sqs.SendMessageInput{
		MessageBody:  aws.String(string(bs)),
		QueueUrl:     aws.String(queueURL.String()),
		DelaySeconds: 0,
	})
	return len(bs), err
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

func infoProvider() io.Reader {
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
	return bytes.NewReader(bs)
}

func stdinProvider() func() io.Reader {
	bs, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("unable to read from stdin: %v", err)
	}
	return func() io.Reader {
		return bytes.NewReader(bs)
	}
}

func fileProvider(fileName string) func() io.Reader {
	bs, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatalf("unable to read file '%s': %v", fileName, err)
	}
	return func() io.Reader {
		return bytes.NewReader(bs)
	}
}

func main() {
	var in func() io.Reader
	var err error

	fileName := flag.String("f", "", "file to send")
	count := flag.Int("c", 1, "send the message this many times")
	interval := flag.Duration("i", time.Millisecond*200, "seconds to wait between sends")
	region := flag.String("region", "local", "AWS region")
	flag.Parse()

	switch *fileName {
	case "":
		in = infoProvider
	case "-":
		in = stdinProvider()
	default:
		in = fileProvider(*fileName)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), func(c *config.LoadOptions) error {
		if *region != "local" {
			c.Region = *region
		}
		return nil
	})
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
		start := time.Now()
		var sentBytes int
		if sentBytes, err = send(context.TODO(), sqsClient, queueURL, in); err != nil {
			log.Fatalf("unable to send message: %v", err)
		}
		diff := time.Now().Sub(start)
		fmt.Printf("%d bytes to %s: time=%d ms\n", sentBytes, queueURL, diff.Milliseconds())
		if i > 0 && i < (*count-1) {
			time.Sleep(*interval)
		}
	}
}
