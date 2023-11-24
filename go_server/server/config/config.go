package config

const (
	LADDR = "127.0.0.1"
	LPORT = "8080"

	SECVAL      = "MyseCretvAlue"
	BEARER      = "Bearer"
	TOKEN_DELIM = "."
	HASH_ALG    = "sha256"
)

type Sconfig struct {
	Server struct {
		InitialChats map[string]string
		Laddr        string
	}
	Token struct {
		TokenDelim string
		SecVal     string
		Bearer     string
		HashAlg    string
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
	cfg.Token.HashAlg = HASH_ALG

	cfg.Server.InitialChats = make(map[string]string)
	chat := cfg.Server.InitialChats
	chat["Echo"] = "Welcome to the `Echo` chat!"
	chat["HCK"] = "Welcome to the `Hack` chat :D"

	return &cfg
}
