package config

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/vzveiteskostrami/goph-keeper/internal/logging"
)

type PConfig struct {
	ServerAddress *string `json:"server_address,omitempty"`
	DatabaseDSN   *string `json:"database_dsn,omitempty"`
	CertAddresses *string `json:"cert_addresses,omitempty"`
}

var cfg *PConfig

func Get() *PConfig {
	return cfg
}

func ReadData() {
	var r PConfig

	cfg = nil

	content, err := os.ReadFile("config.json")
	if err != nil {
		logging.S().Errorln("Ошибка чтения config файла", err)
		return
	}
	err = json.Unmarshal(content, &r)
	if err != nil {
		logging.S().Errorln("Ошибка парсинга config файла", err)
		return
	}

	if r.DatabaseDSN == nil {
		logging.S().Errorln("Не указан database_dsn")
		return
	}

	if r.CertAddresses == nil {
		logging.S().Errorln("Не указан cert_addresses")
		return
	}

	if r.ServerAddress == nil {
		logging.S().Errorln("Не указан server_address", err)
		return
	}

	cfg = &r

	if *r.DatabaseDSN != "" {
		if !strings.Contains(*r.DatabaseDSN, "user") {
			*r.DatabaseDSN += " user=postgres"
		}
		if !strings.Contains(*r.DatabaseDSN, "password") {
			*r.DatabaseDSN += " password=masterkey"
		}
	}
}
