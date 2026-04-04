package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

func LoadDotEnv() error {
	if _, err := os.Stat(".env"); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	return godotenv.Load(".env")
}
