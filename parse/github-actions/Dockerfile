FROM alpine
RUN apk add git
RUN apk add the_silver_searcher
COPY out/github-actions-flag-parser /github-actions-flag-parser
ENTRYPOINT ["/github-actions-flag-parser"]
