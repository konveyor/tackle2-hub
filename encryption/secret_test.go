package encryption

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestSecret(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	secret := Secret{Passphrase: "test"}

	//
	// string
	password := "broker"
	err := secret.Encrypt(&password)
	g.Expect(err).To(gomega.BeNil())
	err = secret.Decrypt(&password)
	g.Expect(err).To(gomega.BeNil())

	//
	// struct
	object := struct {
		Name     string
		User     string `secret:"aes"`
		Password string `secret:"aes"`
		Age      int
		List     []string       `secret:"aes"`
		Map      map[string]any `secret:"aes"`
	}{
		Name:     "elmer",
		User:     "rfudd@warnerbrothers.com",
		Password: "rabbit-slaye",
		Age:      52,
		List: []string{
			"A",
			"B",
			"C",
		},
		Map: map[string]any{
			"k0": "v0",
			"k1": "v1",
			"k2": 2,
		},
	}
	err = secret.Encrypt(&object)
	g.Expect(err).To(gomega.BeNil())
	err = secret.Decrypt(&object)
	g.Expect(err).To(gomega.BeNil())

	//
	// map
	mp := map[string]any{
		"name":     "elmer",
		"user":     "rfudd@warnerbrothers.com",
		"password": "rabbit-slaye",
		"age":      52,
		"list": []string{
			"A",
			"B",
			"C",
		},
	}
	err = secret.Encrypt(mp)
	g.Expect(err).To(gomega.BeNil())
	err = secret.Decrypt(mp)
	g.Expect(err).To(gomega.BeNil())
}
