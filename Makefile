.PHONY: vet

sqs-ping: main.go 
	go build

vet:
	go vet ./...

