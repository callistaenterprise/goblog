package cmd

type Config struct {
	Environment        string `arg:"env:ENVIRONMENT"`
	CockroachdbConnUrl string `arg:"env:COCKROACHDB_CONN_URL"`
	ZipkinServerUrl    string `arg:"env:ZIPKIN_SERVER_URL"`
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
		Environment: "dev",
		CockroachdbConnUrl: "postgresql://cockroachdb1:26257/account?sslmode=disable",
		ZipkinServerUrl:    "http://zipkin:9411",
		ServerConfig: ServerConfig{
			Name: "imageservice",
			Port: "7777",
		},
		AmqpConfig: AmqpConfig{
			ServerUrl: "amqp://guest:guest@rabbitmq:5672/",
		},
	}
}
