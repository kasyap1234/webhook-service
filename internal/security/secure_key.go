//Package security 
// 
package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
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

func GenerateSignature(payload []byte,secretKey string)(string,error){
mac :=hmac.New(sha256.New,[]byte(secretKey))
_,err :=mac.Write(payload)
if err !=nil{
	return "",err 
}
signatureBytes :=mac.Sum(nil)
signature :=hex.EncodeToString(signatureBytes)
return signature, nil
}
