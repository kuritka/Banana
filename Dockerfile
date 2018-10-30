FROM golang:latest as build

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

ARG GH_USER
ARG GH_TOKEN
RUN git config --global credential.helper "/bin/bash -c 'echo username=\$GH_USER; echo password=\$GH_TOKEN'"

WORKDIR /go/src/github.com/Deutsche-Boerse/edt-sftp/
COPY . .

RUN dep ensure

RUN go build -ldflags "-linkmode external -extldflags -static" -a main.go

FROM alpine:latest

RUN apk add ca-certificates

WORKDIR /app/
COPY --from=build /go/src/github.com/Deutsche-Boerse/edt-sftp/main .
COPY conf/sftp_dev.json .
COPY key.private .
CMD ["./main"]