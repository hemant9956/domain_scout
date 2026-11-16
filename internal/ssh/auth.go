package ssh

import (
	"fmt"

	gossh "golang.org/x/crypto/ssh"
)

type AuthInput struct {
	Type       string
	Password   string
	PrivateKey string
	Passphrase string
}

func NewAuth(in AuthInput) ([]gossh.AuthMethod, error) {
	switch in.Type {
	case "password":
		if in.Password == "" {
			return nil, fmt.Errorf("password is required")
		}
		return []gossh.AuthMethod{gossh.Password(in.Password)}, nil
	case "private_key":
		if in.PrivateKey == "" {
			return nil, fmt.Errorf("private_key is required")
		}
		var signer gossh.Signer
		var err error
		if in.Passphrase != "" {
			signer, err = gossh.ParsePrivateKeyWithPassphrase([]byte(in.PrivateKey), []byte(in.Passphrase))
		} else {
			signer, err = gossh.ParsePrivateKey([]byte(in.PrivateKey))
		}
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		return []gossh.AuthMethod{gossh.PublicKeys(signer)}, nil
	default:
		return nil, fmt.Errorf("auth_type must be password or private_key")
	}
}
