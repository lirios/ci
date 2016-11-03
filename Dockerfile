FROM alpine
MAINTAINER Pier Luigi Fiorini <pierluigi.fiorini@gmail.com>
RUN apk add -U ca-certificates && rm -rf /var/cache/apk/*
ADD ci /
ADD web/ /web
RUN sed -i -e 's,ws://localhost:8090/ws,ws://build.liri.io/ws,g' /web/static/js/app.js
CMD ["/ci", "/config.ini"]
EXPOSE 8090
