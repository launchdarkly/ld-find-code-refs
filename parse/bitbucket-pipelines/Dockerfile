FROM alpine:3.8

RUN apk update
RUN apk add --no-cache git
RUN apk add --no-cache the_silver_searcher
RUN apk add --no-cache openssh

COPY bitbucket-pipelines-flag-parser /bitbucket-pipelines-flag-parser

ENTRYPOINT ["/bitbucket-pipelines-flag-parser"]
