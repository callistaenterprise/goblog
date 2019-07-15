FROM alpine:latest

EXPOSE 6868
RUN apk --update add \
    curl \
	tzdata \
	ca-certificates \
	&& rm -rf /var/cache/apk/*

ENV TZ=Europe/Stockholm

ADD bin/vipservice /

HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=5 CMD curl -sSf http://127.0.0.1:6868/health

ENTRYPOINT ["./vipservice"]
