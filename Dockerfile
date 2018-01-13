FROM golang

WORKDIR /go/src/github.com/nksfrank/whisper/src/
COPY . /go/src/github.com/nksfrank/whisper/src

RUN go-wrapper download
RUN go-wrapper install

CMD ["go-wrapper", "run"] # ["app"]