package config

const (
	LADDR = "127.0.0.1"
	LPORT = "8080"

	SECVAL      = "MyseCretvAlue"
	BEARER      = "Bearer"
	TOKEN_DELIM = "."
)

type Sconfig struct {
	Server struct {
		Laddr string
	}
	Token struct {
		TokenDelim string
		SecVal     string
		Bearer     string
	}
	// TODO: log config
}

func New() *Sconfig {
	// TODO: read config from file
	return Default()
}

func Default() *Sconfig {
	cfg := Sconfig{}

	cfg.Server.Laddr = LADDR + ":" + LPORT

	cfg.Token.Bearer = BEARER
	cfg.Token.TokenDelim = TOKEN_DELIM
	cfg.Token.SecVal = SECVAL

	return &cfg
}
