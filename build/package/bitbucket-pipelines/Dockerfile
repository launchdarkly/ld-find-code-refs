FROM alpine:3.8

RUN apk update
RUN apk add --no-cache git

COPY ld-find-code-refs-bitbucket-pipeline /ld-find-code-refs-bitbucket-pipeline

ENTRYPOINT ["/ld-find-code-refs-bitbucket-pipeline"]
