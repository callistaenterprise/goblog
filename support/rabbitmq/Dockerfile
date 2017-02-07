# use rabbitmq official
FROM rabbitmq

# enable management plugin
RUN rabbitmq-plugins enable --offline rabbitmq_management

# enable mqtt plugin
RUN rabbitmq-plugins enable --offline rabbitmq_mqtt

# expose management port
EXPOSE 15672
EXPOSE 5672