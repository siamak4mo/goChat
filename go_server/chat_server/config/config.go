package config

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

const (
	CONF_PATHS = "gochat_server.json:../gochat_server.json:/etc/gochat_server.json"

	LADDR = "127.0.0.1"
	LPORT = "8080"

	ADMIN_LADDR = "127.0.0.1"
	ADMIN_LPORT = "8081"

	SECVAL      = "MyseCretvAlue"
	BEARER      = "Bearer"
	TOKEN_DELIM = "."
	HASH_ALG    = "sha256"
)

type Config struct {
	Token struct {
		Delim   string `json:"token_delim"`
		SecVal  string `json:"token_private_key"`
		Bearer  string `json:"token_bearer"`
		HashAlg string `json:"token_hash_algorithm"`
	}
	Admin struct {
		Addr string `json:"admin_addr"`
	}
	Server struct {
		Addr     string   `json:"listen_addr"`
		Chats    []string `json:"room_names"`
		ChatMOTD []string `json:"room_motds"`
	}
	Log struct {
		LogLevel uint `json:"log_level"`
	}
}

func get_conf_file() (*os.File, string, error) {
	for _, path := range strings.Split(CONF_PATHS, ":") {
		f, err := os.Open(path)

		if err == nil {
			return f, path, nil
		} else if os.IsExist(err) {
			return nil, "", err
		}
	}
	return nil, "", errors.New("Config not found")
}

func New() *Config {
	cfg := Default()
	
	if f, path, err := get_conf_file(); err == nil {
		jp := json.NewDecoder(f)
		if err = jp.Decode(cfg); err != nil {
			println("loading configuration from file failed")
			println("loading default configuration")
			return Default()
		} else {
			println("configuration loaded from " + path)
			return cfg
		}
	}
	println("configuration file does not exist")
	println("loading default configuration")
	return cfg
}

func Default() *Config {
	cfg := Config{}

	cfg.Server.Addr = LADDR + ":" + LPORT
	cfg.Admin.Addr = ADMIN_LADDR + ":" + ADMIN_LPORT

	cfg.Token.Bearer = BEARER
	cfg.Token.Delim = TOKEN_DELIM
	cfg.Token.SecVal = SECVAL
	cfg.Token.HashAlg = HASH_ALG

	cfg.Server.Chats = []string{
		"EcHo",
		"HCK",
	}
	cfg.Server.ChatMOTD = []string{
		"Welcome to the `echo` chat!",
		"Welcome to the `Hack` chat :D",
	}

	cfg.Log.LogLevel = 0 // debug log level

	return &cfg
}
