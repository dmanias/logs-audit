FROM golang:1.18.3-alpine3.16 AS development
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOARCH=amd64 \
    GOOS=linux

WORKDIR /app
EXPOSE 8080
COPY . .
#RUN go mod download
#RUN go test . && go test ./auth
CMD ["go", "run", "./cmd"]