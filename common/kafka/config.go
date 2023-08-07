package kafka

type Config struct {
	Brokers  []string
	Topic    string
	Workers  int
	MaxQueue int
}
