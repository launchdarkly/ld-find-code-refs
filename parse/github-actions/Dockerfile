FROM alpine:3.8

RUN apk add --no-cache git
RUN apk add --no-cache the_silver_searcher

COPY github-actions-flag-parser /github-actions-flag-parser

CMD ["/github-actions-flag-parser"]
