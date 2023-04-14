FROM golang:latest

WORKDIR /dkg-node
COPY . .

RUN rm -rf example.sock
RUN go get -v ./...
RUN go build