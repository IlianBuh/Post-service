package kafka

type Config struct {
	Addrs   []string `json:"addrs"`
	Timeout int      `json:"timeout"`
	Retries int      `json:"retries"`
}
