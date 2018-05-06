FROM iron/base

EXPOSE 7777

ADD dataservice-linux-amd64 /
ADD healthchecker-linux-amd64 /

HEALTHCHECK --interval=10s --timeout=15s CMD ["./healthchecker-linux-amd64", "-port=7070"] || exit 1

ENTRYPOINT ["./dataservice-linux-amd64", "-configServerUrl=http://configserver:8888", "-profile=test", "-configBranch=P13"]
