package config 

const (
	LADDR        = "127.0.0.1"
	LPORT        = "8080"
	LISTEN       = LADDR + LPORT
)

type Sconfig struct {
	Laddr         string
	Token_Diam    string
	// TODO: log config
}


func NewConfig() *Sconfig {
	return &Sconfig{
		Laddr: LADDR + ":" + LPORT,
		Token_Diam: ".",
	}
}
