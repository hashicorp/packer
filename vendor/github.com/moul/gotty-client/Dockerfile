# build
FROM            golang:1.9 as builder
RUN             apt update && apt -y install jq
COPY            . /go/src/github.com/moul/gotty-client
WORKDIR         /go/src/github.com/moul/gotty-client
RUN             make install

# minimal runtime
FROM            scratch
COPY            --from=builder /go/bin/gotty-client /bin/gotty-client
ENTRYPOINT      ["/bin/gotty-client"]
