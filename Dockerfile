FROM golang:1.15-alpine AS build-env
ADD . /src
RUN apk --no-cache add ca-certificates git \
    && cd /src && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /go/bin/ydtb .

FROM python:3.9-alpine
RUN apk add --no-cache ca-certificates ffmpeg \
    && wget https://yt-dl.org/downloads/latest/youtube-dl -O /usr/local/bin/youtube-dl \
    && chmod a+rx /usr/local/bin/youtube-dl
WORKDIR /bot
COPY --from=build-env /go/bin/ydtb ydtb

ENTRYPOINT ["/bot/ydtb"]
