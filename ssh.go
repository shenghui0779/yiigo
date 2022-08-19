package yiigo

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

type SSHKey struct {
	IDRSA       []byte
	IDRSAPub    []byte
	Fingerprint string
}

// GenerateSSHKey returns ssh id_rsa and id_rsa.pub.
// NOTE: id_rsa.pub ends with `\n`
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

// NewSSHIDPubFromPublicKeyBlock returns id_rsa.pub and fingerprint from rsa public key block.
// NOTE: value ends with `\n`
func NewSSHIDPubFromPublicKeyBlock(pemBlock []byte) (idRsaPub []byte, fingerprint string, err error) {
	block, _ := pem.Decode(pemBlock)

	if block == nil {
		err = errors.New("invalid rsa public key")

		return
	}

	pk, err := x509.ParsePKIXPublicKey(block.Bytes)

	if err != nil {
		return
	}

	sshKey, err := ssh.NewPublicKey(pk.(*rsa.PublicKey))

	if err != nil {
		return
	}

	idRsaPub = ssh.MarshalAuthorizedKey(sshKey)
	fingerprint = MD5(string(sshKey.Marshal()))

	return
}

// NewSSHIDPubFromPublicKeyBlock returns id_rsa.pub and fingerprint from rsa public key file.
// NOTE: value ends with `\n`
func NewSSHIDPubFromPublicKeyFile(pemFile string) (idRsaPub []byte, fingerprint string, err error) {
	keyPath, err := filepath.Abs(pemFile)

	if err != nil {
		return
	}

	f, err := os.Open(keyPath)

	if err != nil {
		return
	}

	defer f.Close()

	b, err := ioutil.ReadAll(f)

	if err != nil {
		return
	}

	return NewSSHIDPubFromPublicKeyBlock(b)
}
