package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"strings"
)

const (
	SECVAL = "MyseCretvAlue"
	BEARER = "Bearer"
	T_SEP = "."
)

type Token_t struct{
	Token string
	Signature string
	Username []byte
	hasher hash.Hash
}


func Init_token() *Token_t{  // TODO: config
	t := Token_t{}
	t.hasher = sha256.New()
	return &t;
}
func Init_btoken(bearer_token string) (*Token_t, error){  // TODO: config
	if len(bearer_token) < len(BEARER)+1 ||
		strings.Compare(BEARER, bearer_token[0:len(BEARER)]) != 0{
			return nil, errors.New("Invalid Bearer Token")
		}

	t := Token_t{}
	t.Token = bearer_token[len(BEARER)+1:]
	t.hasher = sha256.New()
	return &t, nil;
}

func (t *Token_t) parse() error{
	token_parts := strings.Split(t.Token, T_SEP)
	if len(token_parts) != 2{
		return errors.New("invalid token")
	}

	uname, err1 := base64.StdEncoding.DecodeString(token_parts[0])
	if err1 != nil{
		return errors.New("invalid token - " + err1.Error())
	}

	t.Username = uname
	t.Signature = token_parts[1]
	return nil
}


func (t *Token_t) MkToken(){
	username_b64 := base64.StdEncoding.EncodeToString(t.Username)
	
	t.hasher.Write(t.Username)
	t.hasher.Write([]byte(SECVAL))
	signature := hex.EncodeToString(t.hasher.Sum(nil))
	t.hasher.Reset()

	t.Token = fmt.Sprintf("%s%s%s", username_b64, T_SEP, signature)
}


func (t *Token_t) Validate() bool{
	e := t.parse()
	if e!=nil{
		return false
	}

	t.hasher.Write(t.Username)
	t.hasher.Write([]byte(SECVAL))
	exp_sign := hex.EncodeToString(t.hasher.Sum(nil))

	if strings.Compare(t.Signature, exp_sign) != 0{
		return false
	}
	return true
}
