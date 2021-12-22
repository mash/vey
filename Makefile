serve:
	cd cmd/vey && go run main.go serve --debug

test:
	go test -timeout 30s -v github.com/mash/vey
	go test -timeout 30s -v github.com/mash/vey/http
