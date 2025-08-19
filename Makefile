# Rinha Backend 2025 - Go Implementation Makefile
.PHONY: build build-docker run test benchmark clean deps fmt lint docker-up docker-down test-endpoints

build:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static" -s -w' -o bin/main ./cmd/main.go

run:
	go run ./cmd/main.go

test-preview:
	cd ./scripts/payment-processor && \
	docker compose up --build -d & \
	cd ./scripts/rinha-test/ && \
	k6 run rinha.js

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

clean:
	rm -rf bin/
	docker compose down -v
	docker system prune -f

deps:
	go mod download
	go mod verify
	go mod tidy

update-deps:
	go get -u ./...
	go mod tidy

fmt:
	go fmt ./...

dev: deps fmt build 

deploy: clean docker-up

