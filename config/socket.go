package config

type Socket struct {
	SocketPath  string `yaml:"socket_path"`
	QueueBuffer int    `yaml:"queue_buffer"`
	Worker      int    `yaml:"worker"`
}
