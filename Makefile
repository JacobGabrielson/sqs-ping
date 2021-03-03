.PHONY: verify

sqs-ping: main.go 
	go build

verify:
	go mod download
	go mod tidy
	go fmt ./...
	go vet ./...

