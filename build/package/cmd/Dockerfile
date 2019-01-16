FROM alpine:3.8

RUN apk update
RUN apk add --no-cache git
RUN apk add --no-cache the_silver_searcher
RUN apk add --no-cache openssh

COPY git-flag-parser /usr/local/bin/git-flag-parser
