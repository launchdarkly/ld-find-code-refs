FROM alpine:3.8

RUN apk update
RUN apk add --no-cache git
RUN apk add --no-cache the_silver_searcher
RUN apk add --no-cache openssh

COPY ld-find-code-refs /usr/local/bin/ld-find-code-refs

ENTRYPOINT ["ld-find-code-refs"]
