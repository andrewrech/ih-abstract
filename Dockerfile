FROM golang:latest
COPY ih-abstract /
ENTRYPOINT ["/polly"]
