FROM iron/base

EXPOSE 7777

RUN apk --update add \
    curl \
	tzdata \
	ca-certificates \
	&& rm -rf /var/cache/apk/*

ENV TZ=Europe/Stockholm

ADD testimages/*.jpg testimages/
ADD bin/imageservice /

HEALTHCHECK --interval=10s --timeout=5s --start-period=10s --retries=5 CMD curl -sSf http://127.0.0.1:7777/health

ENTRYPOINT ["./imageservice", "-configServerUrl=http://configserver:8888", "-profile=test", "-configBranch=P16"]
