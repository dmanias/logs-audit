FROM golang:1.18.3-alpine3.16 AS development
ENV GO111MODULE=on \
    CGO_ENABLED=1  \
    GOARCH=amd64 \
    GOOS=linux

WORKDIR /
EXPOSE 8080
COPY . .
CMD ["go", "run", "."]
