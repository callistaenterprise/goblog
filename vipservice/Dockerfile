FROM iron/base

EXPOSE 6868
ADD vipservice-linux-amd64 /
ADD healthchecker-linux-amd64 /

HEALTHCHECK --interval=3s --timeout=3s CMD ["./healthchecker-linux-amd64", "-port=6868"] || exit 1

ENTRYPOINT ["./vipservice-linux-amd64", "-configServerUrl=http://configserver:8888", "-profile=test", "-configBranch=P9"]