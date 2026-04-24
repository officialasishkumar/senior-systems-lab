.PHONY: fmt vet test race build run probe-http probe-tcp probe-udp docker-build compose-up compose-down e2e

APP := pulsemesh
IMAGE ?= pulsemesh:local

fmt:
	gofmt -w cmd internal

vet:
	go vet ./...

test:
	go test ./...

race:
	go test -race ./...

build:
	go build -trimpath -ldflags="-s -w" -o bin/$(APP) ./cmd/pulsemesh
	go build -trimpath -ldflags="-s -w" -o bin/netprobe ./cmd/netprobe

run:
	go run ./cmd/pulsemesh

probe-http:
	go run ./cmd/netprobe -mode http -addr 127.0.0.1:8080

probe-tcp:
	go run ./cmd/netprobe -mode tcp -addr 127.0.0.1:9090

probe-udp:
	go run ./cmd/netprobe -mode udp -addr 127.0.0.1:9091

docker-build:
	docker build -t $(IMAGE) .

compose-up:
	docker compose up --build

compose-down:
	docker compose down -v

e2e:
	bash tests/e2e/local.sh

