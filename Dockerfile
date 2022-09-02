FROM golang:1.18-bullseye as builder

WORKDIR $GOPATH/src/kong/

COPY . .

RUN go mod download
RUN go mod verify

# Compile the goplug plugin
RUN GOOS=linux GOARCH=amd64 go build -o go-plugins/bin/goplug go-plugins/goplug.go

ARG TC_KONG_IMAGE
FROM ${TC_KONG_IMAGE:-kong:2.8.1}

COPY --from=builder /go/src/kong/go-plugins/bin /usr/local/kong/go-plugins/bin
