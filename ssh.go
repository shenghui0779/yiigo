package yiigo

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"

	"golang.org/x/crypto/ssh"
)

type SSHKey struct {
	IDRSA       []byte
	IDRSAPub    []byte
	Fingerprint string
}

// GenerateSSHKey returns ssh id_rsa and id_rsa.pub.
// Note: id_rsa.pub ends with `\n`
func GenerateSSHKey() (*SSHKey, error) {
	prvKey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		return nil, err
	}

	key := new(SSHKey)

	key.IDRSA = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(prvKey),
	})

	pubKey, err := ssh.NewPublicKey(prvKey.Public())

	if err != nil {
		return nil, err
	}

	key.IDRSAPub = ssh.MarshalAuthorizedKey(pubKey)
	key.Fingerprint = MD5(string(pubKey.Marshal()))

	return key, nil
}

// RSAPemToSSH converts rsa public key from pem to ssh-rsa.
// Note: value ends with `\n`
func RSAPemToSSH(pemPubKey []byte) (sshRSA []byte, fingerprint string, err error) {
	block, _ := pem.Decode(pemPubKey)

	if block == nil {
		err = errors.New("invalid rsa public key")

		return
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)

	if err != nil {
		return
	}

	rsaKey, ok := pubKey.(*rsa.PublicKey)

	if !ok {
		err = errors.New("invalid rsa public key")

		return
	}

	sshKey, err := ssh.NewPublicKey(rsaKey)

	if err != nil {
		return
	}

	sshRSA = ssh.MarshalAuthorizedKey(sshKey)
	fingerprint = MD5(string(sshKey.Marshal()))

	return
}
