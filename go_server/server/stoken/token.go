package stoken

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
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
	Conf      *config.Sconfig
	Username  []byte
}

func New(cfg *config.Sconfig) *Token_t {
	return &Token_t{
		hasher: GetHasher(cfg.Token.HashAlg),
		Conf:   cfg,
	}
}

func New_s(token string, cfg *config.Sconfig) *Token_t {
	return &Token_t{
		Token:  token,
		hasher: GetHasher(cfg.Token.HashAlg),
		Conf:   cfg,
	}
}

func New_b(bearer_token string, cfg *config.Sconfig) (*Token_t, error) {
	if len(bearer_token) < len(cfg.Token.Bearer)+1 ||
		strings.Compare(cfg.Token.Bearer, bearer_token[0:len(cfg.Token.Bearer)]) != 0 {
		return nil, errors.New("Invalid Bearer Token")
	}

	return &Token_t{
		Token:  bearer_token[len(cfg.Token.Bearer)+1:],
		hasher: GetHasher(cfg.Token.HashAlg),
		Conf:   cfg,
	}, nil
}

func GetHasher(hash_name string) hash.Hash {
	if strings.Compare(hash_name, "sha256") == 0 {
		return sha256.New()
	}
	if strings.Compare(hash_name, "sha1") == 0 {
		return sha1.New()
	}
	if strings.Compare(hash_name, "sha512") == 0 {
		return sha512.New()
	}

	return nil
}

func (t *Token_t) parse() error {
	token_parts := strings.Split(t.Token, t.Conf.Token.TokenDelim)
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
	t.hasher.Write([]byte(t.Conf.Token.SecVal))
	signature := hex.EncodeToString(t.hasher.Sum(nil))
	t.hasher.Reset()

	t.Token = fmt.Sprintf("%s%s%s", username_b64, t.Conf.Token.TokenDelim, signature)
}

func (t *Token_t) Validate() bool {
	e := t.parse()
	if e != nil {
		return false
	}

	t.hasher.Write(t.Username)
	t.hasher.Write([]byte(t.Conf.Token.SecVal))
	exp_sign := hex.EncodeToString(t.hasher.Sum(nil))

	if strings.Compare(t.Signature, exp_sign) != 0 {
		return false
	}
	return true
}
