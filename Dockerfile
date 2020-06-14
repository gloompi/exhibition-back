FROM golang:1.14.4-alpine3.12

WORKDIR /app

RUN export GOROOT=$HOME/app
RUN export PATH=$PATH:$GOROOT/bin

COPY go.mod /app
COPY go.sum /app

RUN go mod download

COPY ./ /app

RUN go mod download
RUN go build -o main .

