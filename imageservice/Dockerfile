FROM iron/base

EXPOSE 7777

ADD imageservice-linux-amd64 /
ADD testimages/*.jpg testimages/
ADD healthchecker-linux-amd64 /

HEALTHCHECK --interval=1s --timeout=3s CMD ["./healthchecker-linux-amd64", "-port=7777"] || exit 1

ENTRYPOINT ["./imageservice-linux-amd64", "-configServerUrl=http://configserver:8888", "-profile=test", "-configBranch=nov2017"]
