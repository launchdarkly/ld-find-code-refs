FROM alpine:3.8

RUN apk --update add --no-cache git the_silver_searcher openssh

# ripgrep is not listed on apk, so install it manually
RUN wget https://github.com/BurntSushi/ripgrep/releases/download/0.10.0/ripgrep-0.10.0-x86_64-unknown-linux-musl.tar.gz -O - | tar -xz && cp ripgrep-0.10.0-x86_64-unknown-linux-musl/rg /usr/bin/rg

COPY ld-find-code-refs /usr/local/bin/ld-find-code-refs

ENTRYPOINT ["ld-find-code-refs"]
