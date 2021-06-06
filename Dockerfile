FROM golang:1.16-alpine as builder
COPY . /src
WORKDIR /src
ENV GO111MODULE=on
RUN apk add --no-cache git && CGO_ENABLED=0 GOOS=linux go build -o research-bot

FROM alpine:latest
WORKDIR /root/
RUN apk add --no-cache tzdata ca-certificates
EXPOSE 8080
COPY --from=builder /src/research-bot /usr/bin/
CMD ["/usr/bin/research-bot"]