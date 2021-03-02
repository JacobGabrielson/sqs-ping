.PHONY: vet

sqs-ping: main.go url/url.go
	go build

vet:
	go vet ./...

