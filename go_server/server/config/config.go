package config

const (
	LADDR = "127.0.0.1"
	LPORT = "8080"

	SECVAL      = "MyseCretvAlue"
	BEARER      = "Bearer"
	TOKEN_DELIM = "."
)

type Sconfig struct {
	Laddr      string
	TokenDelim string
	SecVal     string
	Bearer     string
	// TODO: log config
}

func New() *Sconfig {
	return &Sconfig{
		Laddr:      LADDR + ":" + LPORT,
		TokenDelim: TOKEN_DELIM,
	}
}
