package rabbitmq

type Config struct {
	URL        string
	Exchange   string
	Queue      string
	RoutingKey string

	Prefetch int
	PoolSize int
}
