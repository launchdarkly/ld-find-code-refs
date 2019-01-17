FROM alpine:3.8

RUN apk update
RUN apk add --no-cache git
RUN apk add --no-cache the_silver_searcher

COPY ld-find-code-refs-github-action /ld-find-code-refs-github-action

ENTRYPOINT ["/ld-find-code-refs-github-action"]
