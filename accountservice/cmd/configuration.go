package cmd

type Config struct {
	Environment     string `arg:"env:ENVIRONMENT"`
	ZipkinServerUrl string `arg:"env:ZIPKIN_SERVER_URL"`
	ServerConfig
	AmqpConfig
}

type ServerConfig struct {
	Port string `arg:"env:SERVER_PORT"`
	Name string `arg:"env:SERVICE_NAME"`
}

type AmqpConfig struct {
	ServerUrl string `arg:"env:AMQP_SERVER_URL"`
}

func DefaultConfiguration() *Config {
	return &Config{
		Environment:     "dev",
		ZipkinServerUrl: "http://zipkin:9411",
		ServerConfig: ServerConfig{
			Name: "accountservice",
			Port: "6767",
		},
		AmqpConfig: AmqpConfig{
			ServerUrl: "amqp://guest:guest@rabbitmq:5672/",
		},
	}
}
