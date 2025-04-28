package encryption

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestAES(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	aes := New("MyPassphrase")
	plain := "ABCDEFGHIJKLMNOPQUSTUVQXYZ"
	//
	// Encrypt.
	encrypted, err := aes.Encrypt(plain)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(encrypted)).ToNot(gomega.Equal(0))
	//
	// Decrypt.
	decrypted, err := aes.Decrypt(encrypted)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(plain).To(gomega.Equal(decrypted))
}

func TestAESGCM(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	aes := AESGCM{}
	aes.Use("MyPassphrase")
	plain := "ABCDEFGHIJKLMNOPQUSTUVQXYZ"
	//
	// Encrypt.
	encrypted, err := aes.Encrypt(plain)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(encrypted)).ToNot(gomega.Equal(0))
	//
	// Decrypt.
	decrypted, err := aes.Decrypt(encrypted)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(plain).To(gomega.Equal(decrypted))
}
