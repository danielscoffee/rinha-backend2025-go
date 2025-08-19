FROM golang:1.24.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -a -installsuffix cgo \
  -ldflags='-w -s -extldflags "-static"' \
  -gcflags='-l=4' \
  -o main ./cmd/main.go

FROM scratch

# TODO: verify if have certs
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /app/main /main

EXPOSE 8080 

ENTRYPOINT ["/main"]
