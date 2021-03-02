package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

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

func main() {
	flag.String("region", "", "AWS region (defaults to local region)") // TODO: impl
	flag.Parse()

	cfg, err := config.LoadDefaultConfig(context.TODO())
	sqsClient := sqs.NewFromConfig(cfg)

	queueId := flag.Arg(0)

	if err != nil {
		log.Fatalf("unable to load AWS configuration, %v", err)
	}
}
