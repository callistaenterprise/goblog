FROM iron/base

EXPOSE 6666

ADD securityservice-linux-amd64 /
ADD healthchecker-linux-amd64 /

HEALTHCHECK --interval=1s --timeout=3s CMD ["./healthchecker-linux-amd64", "-port=6666"] || exit 1

ENTRYPOINT ["./securityservice-linux-amd64", "-profile=test", "-configBranch=nov2017"]
