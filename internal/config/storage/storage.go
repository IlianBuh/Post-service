package storage

// Storage object representation of json data
type Config struct {
	DBName   string `json:"dbname"`
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Timeout  int    `json:"connection-timeout,omitempty"`
}
