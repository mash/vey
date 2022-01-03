DATE := $(shell date +%Y-%m-%d-%H-%M-%S)
GITVER := $(shell git describe --tags --long --always)

serve:
	cd cmd/vey && go run main.go serve --debug

test:
	go test -timeout 30s -v ./...

test-all:
	go test -timeout 30s -v --tags=aws ./...

#--- lambda related

lambda-all: lambda-clean lambda-build

lambda-clean:
	rm -f cmd/lambda/main

lambda-build:
	cd cmd/lambda && GOOS=linux GOARCH=amd64 go build -ldflags="-s -X \"main.version=$(GITVER)\" -X \"main.builddate=$(DATE)\"" -o ./main main.go
