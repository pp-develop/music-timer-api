FROM golang:1.23.3

RUN go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install github.com/air-verse/air@latest

CMD ["air"]
