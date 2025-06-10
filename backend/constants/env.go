package constants

import (
	"fmt"
	"os"
	"strconv"
)

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func LoadEnv(name string) (string, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return "", fmt.Errorf("env %s not found", name)
	}
	return value, nil
}

func LoadPort() (int, error) {
	portString, err := LoadEnv("PORT")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(portString)
}

func LoadJWTSecret() (string, error) {
	value, err := LoadEnv("JWT_SECRET")
	if err != nil {
		return "", err
	}
	return value, nil
}

func LoadDBConfig() (*DBConfig, error) {
	host, err := LoadEnv("DB_HOST")
	if err != nil {
		return nil, err
	}
	portString, err := LoadEnv("DB_PORT")
	if err != nil {
		return nil, err
	}
	// int型に変換
	port, err := strconv.Atoi(portString)
	if err != nil {
		return nil, err
	}

	user, err := LoadEnv("DB_USER")
	if err != nil {
		return nil, err
	}
	password, err := LoadEnv("DB_PASS")
	if err != nil {
		return nil, err
	}
	dbname, err := LoadEnv("DB_NAME")
	if err != nil {
		return nil, err
	}
	return &DBConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
	}, nil
}
