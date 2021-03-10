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

* `<queue>` can either be the name of the queue or its URL.

### Options

```
  -c int
    	send the message this many times (default 1)
  -command string
    	command to run
  -f string
    	file to send
  -i duration
    	how long to wait between sends (default 200ms)
  -region string
    	AWS region (default "local")
```

* If `-file` is not specified, then a JSON blob will be sent
  containing the hostname and time of day on that host.
* If `-file` is set to `-` then stdin will be read.

### Examples

#### Send default ping 5 times

```
$ sqs-ping -region eu-north-1 -c 5 demo-queue
100 bytes to https://sqs.eu-north-1.amazonaws.com/123456789012/demo-queue: time=206 ms
100 bytes to https://sqs.eu-north-1.amazonaws.com/123456789012/demo-queue: time=205 ms
100 bytes to https://sqs.eu-north-1.amazonaws.com/123456789012/demo-queue: time=206 ms
100 bytes to https://sqs.eu-north-1.amazonaws.com/123456789012/demo-queue: time=205 ms
100 bytes to https://sqs.eu-north-1.amazonaws.com/123456789012/demo-queue: time=203 ms
```

(Note the bytes is only reflective of the SQS `MessageBody` size and
does not include HTTP headers or other arguments to [SQS
SendMessage](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_SendMessage.html)).

#### Send output of systemd-analyze blame

```
$ sqs-ping -command "systemd-analyze blame" demo-queue
11288 bytes to https://sqs.us-west-2.amazonaws.com/123456789012/demo-queue: time=395 ms
```

The message sent will include the following JSON that looks something
like this (`\n` converted to actual newline, and output truncated with
`...` for readability):

```
{
  "Hostname": "labrador.example.com",
  "Timestamp": "Wed, 03 Mar 2021 09:13:39 -0800",
  "Command": "systemd-analyze blame",
  "ExecError": "",
  "Stdout": "10.804s alsatian.service
10.014s blah.service
 8.310s NetworkManager-wait-online.service
 3.618s docker.service
 1.585s postfix@-.service
  866ms ufw.service
  827ms fwupd-refresh.service
  669ms dev-mapper-vg0\\x2d\\x2dcf3dba\\x2droot.device
    ...
    4ms setvtrgb.service
    4ms sys-fs-fuse-connections.mount
    3ms postfix.service
    3ms sys-kernel-config.mount
    2ms motd-news.service
    1ms docker.socket
  851us snapd.socket
",
  "Stderr": ""
}
```

## License

This code is licensed under the [Apache License 2.0](LICENSE.txt).
