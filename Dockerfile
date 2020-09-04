FROM golang
LABEL maintainer="Bogdan Melnik teh.ld86@gmail.com"

ADD . /go/src/github.com/ld86/udp
RUN go get github.com/ld86/udp
RUN go install github.com/ld86/udp

ENTRYPOINT ["/go/bin/udp"]
