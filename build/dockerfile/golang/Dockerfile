FROM golang:1.24.5

RUN go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install github.com/air-verse/air@latest

CMD ["air"]
