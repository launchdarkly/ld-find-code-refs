FROM alpine:3.13.5

RUN apk update
RUN apk add --no-cache git

COPY ../../../dist/ld-find-code-refs-bitbucket-pipeline_linux_amd64/ld-find-code-refs-bitbucket-pipeline /ld-find-code-refs-bitbucket-pipeline

ENTRYPOINT ["/ld-find-code-refs-bitbucket-pipeline"]
