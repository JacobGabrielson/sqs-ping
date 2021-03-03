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
	"os/exec"
	"time"

	"github.com/google/shlex"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func send(ctx context.Context, client *sqs.Client, queueURL *url.URL, reader func() io.Reader) (int, error) {
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
	queueURL, err = url.Parse(id)
	if err == nil && queueURL.IsAbs() {
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
	Command   string
	ExecError string
	Stdout    string
	Stderr    string
	Count     int
}

func infoProvider(command string) func() io.Reader {
	var count int
	hostname, err := os.Hostname()
	if err != nil {
		hostname = fmt.Sprintf("unknown (%s)", err.Error())
	}
	return func() io.Reader {
		var status = localStatus{
			Hostname:  hostname,
			Timestamp: time.Now().Format(time.RFC1123Z),
			Count:     count,
		}
		count++
		if command != "" {
			status.Command = command
			args, err := shlex.Split(command)
			if err != nil {
				status.ExecError = fmt.Sprintf("unable to parse command line: %v", err)
			}
			cmd := exec.Command(args[0], args[1:]...)
			devnull, err := os.Open(os.DevNull)
			if err != nil {
				status.ExecError = fmt.Sprintf("unable to open %s: %v", os.DevNull, err)
			}
			defer devnull.Close()
			cmd.Stdin = devnull
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				status.ExecError = fmt.Sprintf("unable to exec: %v", err)
			} else {
				status.Stdout = string(stdout.Bytes())
				status.Stderr = string(stderr.Bytes())
			}
		}
		bs, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			log.Fatalf("creating info %v", err)
		}
		return bytes.NewReader(bs)
	}
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

// On EC2 instances this will return the default (local) region; on
// other servers will have no effect.
func imdsRegion(ctx context.Context) *string {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil
	}
	client := imds.NewFromConfig(cfg)
	region, err := client.GetRegion(ctx, &imds.GetRegionInput{})
	if err != nil {
		return nil
	}
	return &region.Region
}

func main() {
	var in func() io.Reader
	var err error

	fileName := flag.String("f", "", "file to send")
	count := flag.Int("c", 1, "send the message this many times")
	interval := flag.Duration("i", time.Millisecond*200, "how long to wait between sends")
	region := flag.String("region", "local", "AWS region")
	command := flag.String("command", "", "command to run")
	flag.Parse()

	if *fileName != "" && *command != "" {
		log.Fatalf("cannot specify both -command and -f flags")
	}

	switch *fileName {
	case "":
		in = infoProvider(*command)
	case "-":
		in = stdinProvider()
	default:
		in = fileProvider(*fileName)
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, func(c *config.LoadOptions) error {
		if *region != "local" {
			c.Region = *region
		} else {
			localRegion := imdsRegion(ctx)
			if localRegion != nil {
				c.Region = *localRegion
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("unable to load AWS configuration, %v", err)
	}
	sqsClient := sqs.NewFromConfig(cfg)
	queueId := flag.Arg(0)
	queueURL, err := urlFor(ctx, sqsClient, queueId)
	if err != nil {
		log.Fatalf("unable to find queue URL for '%s': %v", queueId, err)
	}
	for i := 0; i < *count; i++ {
		start := time.Now()
		var sentBytes int
		if sentBytes, err = send(ctx, sqsClient, queueURL, in); err != nil {
			log.Fatalf("unable to send message: %v", err)
		}
		diff := time.Now().Sub(start)
		fmt.Printf("%d bytes to %s: time=%d ms\n", sentBytes, queueURL, diff.Milliseconds())
		if i < (*count - 1) {
			time.Sleep(*interval)
		}
	}
}
