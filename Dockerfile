FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
RUN mkdir /user && \
    echo 'nobody:x:65534:65534:nobody:/:' > /user/passwd && \
    echo 'nobody:x:65534:' > /user/group

COPY ./src /src
WORKDIR /src
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o main .

FROM alpine
COPY --from=builder /src/views /app/views
COPY --from=builder /user/group /user/passwd /etc/
COPY --from=builder /src/main /app/
EXPOSE 8080
USER nobody:nobody
WORKDIR /app
ENTRYPOINT ["./main"]