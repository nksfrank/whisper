FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
RUN adduser -D -g '' whisperuser
COPY ./src /src
WORKDIR /src
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/whisper

FROM scratch
COPY --from=builder /src/views /go/bin/whisper/views
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/whisper /go/bin/whisper
USER whisperuser

EXPOSE 80
ENTRYPOINT ["/go/bin/whisper"]