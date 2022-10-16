FROM golang:1.18.3-alpine3.16 AS development
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOARCH=amd64 \
    GOOS=linux

WORKDIR /app
EXPOSE 8080
COPY . .
RUN go mod download
RUN go curl ./cmd/... ./auth/...
CMD ["go", "run", "./cmd"]

FROM golang:1.18.3-alpine3.16 AS build
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOARCH=amd64 \
    GOOS=linux

COPY --from=development /app/ /app/
WORKDIR  /app
RUN go build -o logs-audit

FROM alpine:3.16 AS production
COPY --from=build /app/logs-audit /usr/local/logs-audit
EXPOSE 8080
USER nobody:nobody

ENTRYPOINT ["/usr/local/logs-audit"]