sqs-ping sends data to an SQS queue, in a manner vaguely resembling
the [ping](https://en.wikipedia.org/wiki/Ping_(networking_utility))
command.

## Installation

```
git clone https://github.com/JacobGabrielson/sqs-ping.git
cd sqs-ping
make
```

## Usage

```
sqs-ping [options] <queue>
```

Options include:

```
  -c int
    	send the message this many times (default 1)
  -f string
    	file to send
  -i duration
    	seconds to wait between sends (default 200ms)
  -region string
    	AWS region (default "local")
```

* `<queue>` can either be the name of the queue or its URL.
* If `-file` is not specified, then a JSON blob will be sent
  containing the hostname and time of day on that host.
* If `-file` is set to `-` then stdin will be read.

### Example

```
./sqs-ping -region eu-north-1 -c 5 demo-queue
100 bytes to https://sqs.eu-north-1.amazonaws.com/123456789012/demo-queue: time=206 ms
100 bytes to https://sqs.eu-north-1.amazonaws.com/123456789012/demo-queue: time=205 ms
100 bytes to https://sqs.eu-north-1.amazonaws.com/123456789012/demo-queue: time=206 ms
100 bytes to https://sqs.eu-north-1.amazonaws.com/123456789012/demo-queue: time=205 ms
100 bytes to https://sqs.eu-north-1.amazonaws.com/123456789012/demo-queue: time=203 ms
```

(Note the bytes is only reflective of the SQS `MessageBody` size and
does not include HTTP headers or other arguments to [SQS
SendMessage](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_SendMessage.html)).

## License

This code is licensed under the [Apache License 2.0](LICENSE.txt).
