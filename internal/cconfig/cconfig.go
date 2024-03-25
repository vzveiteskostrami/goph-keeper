package cconfig

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

type Configc struct {
	Login         *string `json:"login,omitempty"`
	Operation     *string
	ServerAddress *string `json:"server_address,omitempty"`
}

var cfg Configc

func ReadData() error {
	content, err := os.ReadFile("config.json")
	if err != nil {
		return fmt.Errorf("ошибка чтения config файла: " + err.Error())
	}

	err = json.Unmarshal(content, &cfg)
	if err != nil {
		return fmt.Errorf("ошибка парсинга config файла: " + err.Error())
	}

	if cfg.ServerAddress == nil || *cfg.ServerAddress == "" {
		fmt.Println()
		return fmt.Errorf("в config.json не указан server_address")
	}

	s := *cfg.ServerAddress
	if s[len(s)-1] != []byte("/")[0] {
		*cfg.ServerAddress += "/"
	}

	lg := flag.String("l", "Логин не указан в командной строке", "User's login")
	cfg.Operation = flag.String("o", "", "User's operation")

	flag.Parse()

	if *lg != "Логин не указан в командной строке" {
		cfg.Login = lg
	}
	return nil
}

func Get() Configc {
	return cfg
}
