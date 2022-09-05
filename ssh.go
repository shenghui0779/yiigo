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
	IDRsa       []byte
	IDRsaPub    []byte
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

	key.IDRsa = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(prvKey),
	})

	pubKey, err := ssh.NewPublicKey(prvKey.Public())

	if err != nil {
		return nil, err
	}

	key.IDRsaPub = ssh.MarshalAuthorizedKey(pubKey)
	key.Fingerprint = MD5(string(pubKey.Marshal()))

	return key, nil
}

// NewIDRsaPubFromPemBlock returns ssh id_rsa.pub and fingerprint from rsa public key (pem block).
// NOTE: value ends with `\n`
func NewIDRsaPubFromPemBlock(pemBlock []byte) (idRsaPub []byte, fingerprint string, err error) {
	block, _ := pem.Decode(pemBlock)

	if block == nil {
		return nil, "", errors.New("no PEM data is found")
	}

	pk, err := x509.ParsePKIXPublicKey(block.Bytes)

	if err != nil {
		return nil, "", err
	}

	sshKey, err := ssh.NewPublicKey(pk.(*rsa.PublicKey))

	if err != nil {
		return nil, "", err
	}

	idRsaPub = ssh.MarshalAuthorizedKey(sshKey)
	fingerprint = MD5(string(sshKey.Marshal()))

	return
}

// NewIDRsaPubFromPemFile returns ssh id_rsa.pub and fingerprint from rsa public key (pem file).
// NOTE: value ends with `\n`
func NewIDRsaPubFromPemFile(pemFile string) (idRsaPub []byte, fingerprint string, err error) {
	keyPath, err := filepath.Abs(pemFile)

	if err != nil {
		return nil, "", err
	}

	f, err := os.Open(keyPath)

	if err != nil {
		return nil, "", err
	}

	defer f.Close()

	b, err := ioutil.ReadAll(f)

	if err != nil {
		return nil, "", err
	}

	return NewIDRsaPubFromPemBlock(b)
}
