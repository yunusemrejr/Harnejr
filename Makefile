.PHONY: test build daemon web clean

test:
	go test ./...

build: daemon

daemon:
	mkdir -p bin
	go build -o bin/harnejrd ./cmd/harnejrd

web:
	pnpm --filter @harnejr/web build

clean:
	rm -rf bin apps/web/dist packages/*/dist
