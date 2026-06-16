package security

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateSecureKey() (string, error) {
	secretBytes :=make([]byte,32)
	_,err :=rand.Read(secretBytes)
	if err !=nil{
		return "",err 
	}
	webhookSecret :=hex.EncodeToString(secretBytes)
	return webhookSecret, nil
}


