FROM golang:1.9.6 as build

RUN mkdir -p /go/src/github.com/openfaas/proxy
WORKDIR /go/src/github.com/openfaas/proxy

COPY main.go    .
COPY reverse.go .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /usr/bin/proxy .

FROM alpine:3.7

RUN addgroup -S app \
    && adduser -S -g app app

WORKDIR /home/app

EXPOSE 8080
ENV http_proxy      ""
ENV https_proxy     ""

COPY --from=build /usr/bin/proxy    .

RUN chown -R app:app ./

USER app

CMD ["./proxy"]
