FROM ofayau/ejre:8-jre

EXPOSE 8282

ADD turbine-amqp-executable-2.0.0-DP.3.jar app.jar

ENTRYPOINT ["java", "-Damqp.broker.url=amqp://guest:guest@rabbitmq:5672", "-Djava.security.egd=file:/dev/./urandom", "-jar", "app.jar"]
