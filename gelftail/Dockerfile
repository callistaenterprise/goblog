FROM iron/base

EXPOSE 12202/udp
ADD gelftail-linux-amd64 /
ADD token.txt /

ENTRYPOINT ["./gelftail-linux-amd64", "-port=12202"]