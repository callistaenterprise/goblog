FROM ofayau/ejre:8-jre

MAINTAINER Magnus Larsson <magnus.larsson.ml@gmail.com>

EXPOSE 8080

ADD app.jar /
ADD healthchecker-linux-amd64 /

HEALTHCHECK --retries=10 --interval=3s --timeout=3s CMD ["./healthchecker-linux-amd64", "-port=8080"] || exit 1

ENTRYPOINT ["java","-Djava.security.egd=file:/dev/./urandom","-jar","/app.jar"]
