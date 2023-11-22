package stoken

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"server/server/config"
	"strings"
)

type Token_t struct {
	Token     string
	Signature string
	hasher    hash.Hash
	Conf      config.Sconfig
	Username  []byte
}

func New(cfg config.Sconfig) *Token_t {
	t := Token_t{}
	t.hasher = sha256.New()
	t.Conf = cfg
	return &t
}

func New_s(token string, cfg config.Sconfig) *Token_t {
	t := Token_t{}
	t.Token = token
	t.hasher = sha256.New()
	t.Conf = cfg
	return &t
}

func New_b(bearer_token string, cfg config.Sconfig) (*Token_t, error) {
	if len(bearer_token) < len(cfg.Bearer)+1 ||
		strings.Compare(cfg.Bearer, bearer_token[0:len(cfg.Bearer)]) != 0 {
		return nil, errors.New("Invalid Bearer Token")
	}

	t := Token_t{}
	t.Token = bearer_token[len(cfg.Bearer)+1:]
	t.hasher = sha256.New()
	t.Conf = cfg
	return &t, nil
}

func (t *Token_t) parse() error {
	token_parts := strings.Split(t.Token, t.Conf.TokenDelim)
	if len(token_parts) != 2 {
		return errors.New("invalid token")
	}

	uname, err1 := base64.StdEncoding.DecodeString(token_parts[0])
	if err1 != nil {
		return errors.New("invalid token - " + err1.Error())
	}

	t.Username = uname
	t.Signature = token_parts[1]
	return nil
}

func (t *Token_t) MkToken() {
	username_b64 := base64.StdEncoding.EncodeToString(t.Username)

	t.hasher.Write(t.Username)
	t.hasher.Write([]byte(t.Conf.SecVal))
	signature := hex.EncodeToString(t.hasher.Sum(nil))
	t.hasher.Reset()

	t.Token = fmt.Sprintf("%s%s%s", username_b64, t.Conf.TokenDelim, signature)
}

func (t *Token_t) Validate() bool {
	e := t.parse()
	if e != nil {
		return false
	}

	t.hasher.Write(t.Username)
	t.hasher.Write([]byte(t.Conf.SecVal))
	exp_sign := hex.EncodeToString(t.hasher.Sum(nil))

	if strings.Compare(t.Signature, exp_sign) != 0 {
		return false
	}
	return true
}
