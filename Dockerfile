FROM golang:apline AS builder
WORKDIR /src
COPY ./src /src

RUN go-wrapper download
RUN go-wrapper install
RUN go build -o goapp

FROM alpine
WORKDIR /app
COPY --from=builder /src/goapp /app

EXPOSE 80
ENTRYPOINT ./goapp