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
Usage of ./sqs-ping:
  -c int
    	send the message this many times (default 1)
  -f string
    	file to send
  -i duration
    	seconds to wait between sends (default 200ms)
  -region string
    	AWS region (default "local")
```

## License

This code is licensed under the [Apache License 2.0](LICENSE.txt).
