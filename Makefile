.PHONY: test build daemon web clean install doctor

test:
	go test ./...

build: daemon web

daemon:
	mkdir -p bin
	go build -o bin/harnejrd ./cmd/harnejrd

web:
	pnpm --filter @harnejr/web build

install:
	bash install.sh

doctor:
	scripts/doctor.sh

clean:
	rm -rf bin apps/web/dist packages/*/dist
