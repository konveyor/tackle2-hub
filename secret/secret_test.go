package secret

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestSecret(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

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
	g.Expect(err).To(gomega.BeNil())
	err = secret.Decrypt(&s)
	g.Expect(err).To(gomega.BeNil())
	g.Expect("hello").To(gomega.Equal(s))

	//
	// different passphrase
	cipher2 := AESGCM{}
	cipher2.Use("other")
	secret2 := Secret{Cipher: &cipher2}
	s2 := "hello"
	err = secret.Encrypt(&s2)
	g.Expect(err).To(gomega.BeNil())
	err = secret2.Decrypt(&s2)
	g.Expect(err).ToNot(gomega.BeNil())

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
	g.Expect(err).To(gomega.BeNil())
	g.Expect(name).To(gomega.Equal(object.Name))
	g.Expect(user).ToNot(gomega.Equal(object.User))
	g.Expect(password).ToNot(gomega.Equal(object.Password))
	err = secret.Decrypt(&object)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(name).To(gomega.Equal(object.Name))
	g.Expect(user).To(gomega.Equal(object.User))
	g.Expect(password).To(gomega.Equal(object.Password))

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
	g.Expect(err).To(gomega.BeNil())
	g.Expect(name).ToNot(gomega.Equal(mp["name"]))
	g.Expect(user).ToNot(gomega.Equal(mp["user"]))
	g.Expect(password).ToNot(gomega.Equal(mp["password"]))
	g.Expect(a).ToNot(gomega.Equal(mp["list"].([]string)[0]))
	err = secret.Decrypt(mp)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(name).To(gomega.Equal(mp["name"]))
	g.Expect(user).To(gomega.Equal(mp["user"]))
	g.Expect(password).To(gomega.Equal(mp["password"]))
	g.Expect(a).To(gomega.Equal(mp["list"].([]string)[0]))
}

func TestPackage(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	name := "elmer"
	user := "rfudd@warnerbrothers.com"
	password := "rabbit-slayer"

	//
	// string
	s := "hello"
	err := Encrypt(&s)
	g.Expect(err).To(gomega.BeNil())
	err = Decrypt(&s)
	g.Expect(err).To(gomega.BeNil())
	g.Expect("hello").To(gomega.Equal(s))

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
	g.Expect(err).To(gomega.BeNil())
	g.Expect(name).To(gomega.Equal(object.Name))
	g.Expect(user).ToNot(gomega.Equal(object.User))
	g.Expect(password).ToNot(gomega.Equal(object.Password))
	err = Decrypt(&object)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(name).To(gomega.Equal(object.Name))
	g.Expect(user).To(gomega.Equal(object.User))
	g.Expect(password).To(gomega.Equal(object.Password))

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
	g.Expect(err).To(gomega.BeNil())
	g.Expect(name).ToNot(gomega.Equal(mp["name"]))
	g.Expect(user).ToNot(gomega.Equal(mp["user"]))
	g.Expect(password).ToNot(gomega.Equal(mp["password"]))
	g.Expect(a).ToNot(gomega.Equal(mp["list"].([]string)[0]))
	err = Decrypt(mp)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(name).To(gomega.Equal(mp["name"]))
	g.Expect(user).To(gomega.Equal(mp["user"]))
	g.Expect(password).To(gomega.Equal(mp["password"]))
	g.Expect(a).To(gomega.Equal(mp["list"].([]string)[0]))
}
