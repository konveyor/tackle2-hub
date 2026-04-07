package secret

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestSecret(t *testing.T) {
	g := NewGomegaWithT(t)

	cipher := AESGCM{}
	cipher.Use("test")
	secret := Secret{Cipher: &cipher}
	name := "elmer"
	user := "rfudd@warnerbrothers.com"
	password := "rabbit-slayer"

	//
	// string
	s := "hello"
	err := secret.Encrypt(&s)
	g.Expect(err).To(BeNil())
	err = secret.Decrypt(&s)
	g.Expect(err).To(BeNil())
	g.Expect("hello").To(Equal(s))

	//
	// different passphrase
	cipher2 := AESGCM{}
	cipher2.Use("other")
	secret2 := Secret{Cipher: &cipher2}
	s2 := "hello"
	err = secret.Encrypt(&s2)
	g.Expect(err).To(BeNil())
	err = secret2.Decrypt(&s2)
	g.Expect(err).ToNot(BeNil())

	//
	// struct
	object := struct {
		Name     string
		User     string `secret:""`
		Password string `secret:""`
		Age      int
		List     []string       `secret:""`
		Map      map[string]any `secret:""`
	}{
		Name:     name,
		User:     user,
		Password: password,
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
	g.Expect(err).To(BeNil())
	g.Expect(name).To(Equal(object.Name))
	g.Expect(user).ToNot(Equal(object.User))
	g.Expect(password).ToNot(Equal(object.Password))
	err = secret.Decrypt(&object)
	g.Expect(err).To(BeNil())
	g.Expect(name).To(Equal(object.Name))
	g.Expect(user).To(Equal(object.User))
	g.Expect(password).To(Equal(object.Password))

	//
	// map
	a := "A"
	mp := map[string]any{
		"name":     "elmer",
		"user":     user,
		"password": password,
		"age":      52,
		"list": []string{
			a,
			"B",
			"C",
		},
	}
	err = secret.Encrypt(mp)
	g.Expect(err).To(BeNil())
	g.Expect(name).ToNot(Equal(mp["name"]))
	g.Expect(user).ToNot(Equal(mp["user"]))
	g.Expect(password).ToNot(Equal(mp["password"]))
	g.Expect(a).ToNot(Equal(mp["list"].([]string)[0]))
	err = secret.Decrypt(mp)
	g.Expect(err).To(BeNil())
	g.Expect(name).To(Equal(mp["name"]))
	g.Expect(user).To(Equal(mp["user"]))
	g.Expect(password).To(Equal(mp["password"]))
	g.Expect(a).To(Equal(mp["list"].([]string)[0]))
}

func TestPackage(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	name := "elmer"
	user := "rfudd@warnerbrothers.com"
	password := "rabbit-slayer"

	//
	// string
	s := "hello"
	err := Encrypt(&s)
	g.Expect(err).To(BeNil())
	err = Decrypt(&s)
	g.Expect(err).To(BeNil())
	g.Expect("hello").To(Equal(s))

	//
	// struct
	object := struct {
		Name     string
		User     string `secret:""`
		Password string `secret:""`
		Age      int
		List     []string       `secret:""`
		Map      map[string]any `secret:""`
	}{
		Name:     name,
		User:     user,
		Password: password,
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
	err = Encrypt(&object)
	g.Expect(err).To(BeNil())
	g.Expect(name).To(Equal(object.Name))
	g.Expect(user).ToNot(Equal(object.User))
	g.Expect(password).ToNot(Equal(object.Password))
	err = Decrypt(&object)
	g.Expect(err).To(BeNil())
	g.Expect(name).To(Equal(object.Name))
	g.Expect(user).To(Equal(object.User))
	g.Expect(password).To(Equal(object.Password))

	//
	// map
	a := "A"
	mp := map[string]any{
		"name":     "elmer",
		"user":     user,
		"password": password,
		"age":      52,
		"list": []string{
			a,
			"B",
			"C",
		},
	}
	err = Encrypt(mp)
	g.Expect(err).To(BeNil())
	g.Expect(name).ToNot(Equal(mp["name"]))
	g.Expect(user).ToNot(Equal(mp["user"]))
	g.Expect(password).ToNot(Equal(mp["password"]))
	g.Expect(a).ToNot(Equal(mp["list"].([]string)[0]))
	err = Decrypt(mp)
	g.Expect(err).To(BeNil())
	g.Expect(name).To(Equal(mp["name"]))
	g.Expect(user).To(Equal(mp["user"]))
	g.Expect(password).To(Equal(mp["password"]))
	g.Expect(a).To(Equal(mp["list"].([]string)[0]))
}

// TestHashSecretDeterministic tests that hashing is deterministic.
func TestHashSecretDeterministic(t *testing.T) {
	g := NewGomegaWithT(t)

	secret := "test-secret-key"
	hash1 := Hash(secret)
	hash2 := Hash(secret)
	g.Expect(hash1).To(Equal(hash2))
	hash3 := Hash("different-secret")
	g.Expect(hash1).NotTo(Equal(hash3))
	g.Expect(hash1).NotTo(BeEmpty())
	hash4 := Hash(hash1)
	g.Expect(hash4).To(Equal(hash2))
}
