FROM alpine:3.8

RUN apk update
RUN apk add --no-cache git the_silver_searcher

# ripgrep is not listed on apk, so install it manually
RUN wget https://github.com/BurntSushi/ripgrep/releases/download/0.10.0/ripgrep-0.10.0-x86_64-unknown-linux-musl.tar.gz -O - | tar -xz && cp ripgrep-0.10.0-x86_64-unknown-linux-musl/rg /usr/bin/rg

COPY ld-find-code-refs-github-action /ld-find-code-refs-github-action

ENTRYPOINT ["/ld-find-code-refs-github-action"]
