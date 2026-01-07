.PHONY: run test build clean

run:
	CGO_ENABLED=0 go run ./cmd/bot/

test:
	CGO_ENABLED=0 go test ./... -v

build:
	CGO_ENABLED=0 go build -o bin/bot ./cmd/bot/

clean:
	rm -rf bin/

