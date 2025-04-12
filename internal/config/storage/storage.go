package storage

// Storage object representation of json data
type Storage struct {
	DBName   string `json:"dbname"`
	User     string `json:"name"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Timeout  int    `json:"connection-timeout,omitempty"`
}
