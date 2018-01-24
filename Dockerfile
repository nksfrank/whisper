FROM golang

WORKDIR /go/src/github.com/nksfrank/whisper/src/
COPY ./src /go/src/github.com/nksfrank/whisper/src

RUN go-wrapper download
RUN go-wrapper install

EXPOSE 80

CMD ["go-wrapper", "run"] # ["app"]