serve:
	cd cmd/vey && go run main.go serve

test:
	go test -timeout 30s -v ./...
