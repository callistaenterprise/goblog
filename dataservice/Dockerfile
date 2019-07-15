FROM alpine:latest

EXPOSE 7070

RUN apk --update add \
    curl \
	tzdata \
	ca-certificates \
	&& rm -rf /var/cache/apk/*

ENV TZ=Europe/Stockholm

ADD bin/dataservice /

HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=5 CMD curl -sSf http://127.0.0.1:7070/health

ENTRYPOINT ["./dataservice"]
