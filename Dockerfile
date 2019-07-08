FROM golang:1.11.4
COPY . /go/src/github.com/maddevsio/mad-internship-bot/
WORKDIR /go/src/github.com/maddevsio/mad-internship-bot
RUN make install_dependencies
RUN GOOS=linux GOARCH=amd64 go build -o mad-internship-bot main.go

FROM debian:9.8
LABEL maintainer="Anatoliy Fedorenko <fedorenko.tolik@gmail.com>"
RUN  apt-get update \
  && apt-get install -y --no-install-recommends ca-certificates locales wget \
  && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
RUN localedef -i en_US -c -f UTF-8 -A /usr/share/locale/locale.alias en_US.UTF-8
ENV LANG en_US.utf8

COPY active.en.toml  /
COPY active.ru.toml  / 
COPY --from=0 /go/src/github.com/maddevsio/mad-internship-bot/mad-internship-bot /
COPY --from=0 /go/src/github.com/maddevsio/mad-internship-bot/migrations /migrations
COPY --from=0 /go/src/github.com/maddevsio/mad-internship-bot/goose /
COPY entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]