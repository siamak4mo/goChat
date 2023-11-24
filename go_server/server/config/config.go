package config

const (
	LADDR = "127.0.0.1"
	LPORT = "8080"

	SECVAL      = "MyseCretvAlue"
	BEARER      = "Bearer"
	TOKEN_DELIM = "."
	HASH_ALG    = "sha256"
)

type Config struct {
	Token struct {
		TokenDelim string
		SecVal     string
		Bearer     string
		HashAlg    string
	}
	Server struct {
		Laddr        string
		IChats       []string
		IChatBanners []string
	}
	// TODO: log config
}

func New() *Config {
	// TODO: read config from file
	return Default()
}

func Default() *Config {
	cfg := Config{}

	cfg.Server.Laddr = LADDR + ":" + LPORT

	cfg.Token.Bearer = BEARER
	cfg.Token.TokenDelim = TOKEN_DELIM
	cfg.Token.SecVal = SECVAL
	cfg.Token.HashAlg = HASH_ALG

	cfg.Server.IChats = []string{
		"EcHo",
		"HCK",
	}
	cfg.Server.IChatBanners = []string{
		"Welcome to the `echo` chat!",
		"Welcome to the `Hack` chat :D",
	}

	return &cfg
}
