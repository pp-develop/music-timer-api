FROM golang:1.21.4

RUN go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install github.com/cosmtrek/air@latest

CMD ["air"]
