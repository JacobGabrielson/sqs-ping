package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/sqs/goreturns/returns"
)

func idToURL(id string) (queueURL *url.URL, err error) {
	if len(id) == 0 {
		return nil, fmt.Errorf("empty identifier")
	}
	if queueURL, err = url.Parse(id); err != nil {
		return
	}

}

func main() {
	flag.Parse()

	queueId := flag.Arg(0)

	_, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		log.Fatalf("unable to load AWS configuration, %v", err)
	}
}
