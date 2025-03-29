FROM golang:1.24 as builder

WORKDIR $GOPATH/src/github.com/mark-adams/gcp-ip-list
COPY . .

RUN go get -v ./...
RUN GOOS=linux CGO_ENABLED=0 go build -o /go/bin/gcp-ip-list github.com/mark-adams/gcp-ip-list/cmd/gcp-ip-list

FROM cgr.dev/chainguard/static

COPY --from=builder /go/bin/gcp-ip-list /bin/gcp-ip-list

ENTRYPOINT ["/bin/gcp-ip-list"]