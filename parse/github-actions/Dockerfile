FROM iron/go:1.10

RUN apk add --no-cache git
RUN apk add --no-cache the_silver_searcher

COPY github-actions-flag-parser /github-actions-flag-parser
ENTRYPOINT ["/github-actions-flag-parser"]
