FROM golang:1.21.7-alpine3.19 as builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN GOOS=linux GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 \
    go build -o client ./cmd/client/main.go
RUN GOOS=linux GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 \
    go build -o server ./cmd/server/main.go

FROM alpine:3.19

COPY --from=builder src/client /app/client
COPY --from=builder src/server /app/server

CMD server
