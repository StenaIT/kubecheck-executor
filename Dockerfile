FROM golang:alpine as builder
RUN adduser -D -u 10001 scratchuser
RUN apk add --no-cache git ca-certificates && update-ca-certificates
COPY . $GOPATH/src/kubecheck-executor/
WORKDIR $GOPATH/src/kubecheck-executor/
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/kubecheck-executor

FROM scratch
COPY --from=builder /etc/passwd /etc/passwd
USER scratchuser
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/kubecheck-executor /go/bin/kubecheck-executor
EXPOSE 8113
ENTRYPOINT ["/go/bin/kubecheck-executor"]
