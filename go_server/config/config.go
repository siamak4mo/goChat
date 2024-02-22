package config

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

type ConFunc func(*Config)

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
	ConfigPath string
}

func fileof(path string) (f *os.File, err error) {
	f, err = os.Open(path)
	if err == nil {
		return
	} else if os.IsExist(err) {
		return nil, errors.New("file is not readable")
	}
	return nil, err
}

func get_conf_file(config_path ...string) (*os.File, string, error) {
	for _, path := range config_path {
		f, err := fileof(path)
		if err == nil {
			return f, path, err
		}
	}
	for _, path := range strings.Split(CONF_PATHS, ":") {
		f, err := fileof(path)
		if err == nil {
			return f, path, err
		}
	}
	return nil, "", errors.New("Config File Not Found")
}

func (c *Config) load_config() {
	f, path, err := get_conf_file(c.ConfigPath)
	if err == nil {
		jp := json.NewDecoder(f)
		if err = jp.Decode(c); err != nil {
			println("loading configuration failed -- " + err.Error())
			println("loading default configuration")
		} else {
			println("configuration loaded from " + path)
		}
	} else {
		println("loading configuration failed -- " + err.Error())
		println("using the default configuration")
	}
}

func WithConfigPath(path string) ConFunc {
	return func(c *Config) {
		defer c.load_config()
		c.ConfigPath = path
	}
}

func New(config_funcs ...ConFunc) *Config {
	cfg := Default()
	for _, fun := range config_funcs {
		fun(cfg)
	}

	/* load config from default paths if it's not being loaded already */
	if len(cfg.ConfigPath) == 0 {
		cfg.load_config()
	}
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
