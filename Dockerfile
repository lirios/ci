FROM alpine
MAINTAINER Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
RUN apk add -U ca-certificates && rm -rf /var/cache/apk/*
ADD ci /
ADD web/ /web
CMD ["/ci", "/config.ini"]
EXPOSE 8090
