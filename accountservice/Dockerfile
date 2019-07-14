FROM alpine:latest

EXPOSE 6767

RUN apk --update add \
    curl \
	tzdata \
	ca-certificates \
	&& rm -rf /var/cache/apk/*

ENV TZ=Europe/Stockholm

ADD bin/accountservice /

HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=5 CMD curl -sSf http://127.0.0.1:6767/health
ENTRYPOINT ["./accountservice", "-configServerUrl=http://configserver:8888", "-profile=test", "-configBranch=P16"]
