# Dockerfile to create a go-getter container with smbclient dependency that is used by the get_smb.go tests
FROM golang:latest

COPY . /go-getter
WORKDIR /go-getter

RUN go mod download
RUN apt-get update
RUN apt-get -y install smbclient
