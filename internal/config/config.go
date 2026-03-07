package config

type Config struct {
	PostgresDSN string
	RedisAddr   string
	NatsURL     string
}

func Load() *Config {

	return &Config{
		PostgresDSN: "postgres://admin:admin@postgres:5432/shortener?sslmode=disable",
		RedisAddr:   "redis:6379",
		NatsURL:     "nats://nats:4222",
	}
}
