FROM alpine:3.22.2

RUN apk update
RUN apk add --no-cache git
RUN apk add --no-cache openssh

COPY ld-find-code-refs /usr/local/bin/ld-find-code-refs

ENTRYPOINT ["ld-find-code-refs"]
