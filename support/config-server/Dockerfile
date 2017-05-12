FROM davidcaste/alpine-java-unlimited-jce

EXPOSE 8888

ADD ./build/libs/*.jar app.jar
ADD ./server.jks /

# Regarding settings of java.security.egd, see http://wiki.apache.org/tomcat/HowTo/FasterStartUp#Entropy_Source
ENTRYPOINT ["java","-Dspring.profiles.active=docker","-Djava.security.egd=file:/dev/./urandom","-jar","/app.jar"]
