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

// TestHashPassword tests password hashing functionality.
func TestHashPassword(t *testing.T) {
	g := NewGomegaWithT(t)

	password := "MySecurePassword123!"

	// Hash the password
	hashed, err := HashPassword(password)
	g.Expect(err).To(BeNil())
	g.Expect(hashed).NotTo(BeEmpty())
	g.Expect(hashed).NotTo(Equal(password))

	// Verify hash starts with bcrypt prefix
	g.Expect(hashed).To(HavePrefix("bcrypt:"))

	// Hash should be base64 encoded (80 chars for prefix + base64 of 60-byte bcrypt)
	g.Expect(len(hashed)).To(BeNumerically(">", 60))

	// Hashing same password produces different hash (due to salt)
	hashed2, err := HashPassword(password)
	g.Expect(err).To(BeNil())
	g.Expect(hashed2).NotTo(Equal(hashed))
}

// TestHashPasswordDoubleHashPrevention tests that already-hashed passwords are not re-hashed.
func TestHashPasswordDoubleHashPrevention(t *testing.T) {
	g := NewGomegaWithT(t)

	password := "MySecurePassword123!"

	// Hash the password
	hashed1, err := HashPassword(password)
	g.Expect(err).To(BeNil())

	// Try to hash the already-hashed password
	hashed2, err := HashPassword(hashed1)
	g.Expect(err).To(BeNil())
	g.Expect(hashed2).To(Equal(hashed1))
}

// TestHashPasswordEmptyString tests handling of empty password.
func TestHashPasswordEmptyString(t *testing.T) {
	g := NewGomegaWithT(t)

	hashed, err := HashPassword("")
	g.Expect(err).To(BeNil())
	g.Expect(hashed).To(Equal(""))
}

// TestHashPasswordStartingWithDollar2 tests that passwords starting with $2 are hashed correctly.
func TestHashPasswordStartingWithDollar2(t *testing.T) {
	g := NewGomegaWithT(t)

	// Password that starts with $2 but isn't a bcrypt hash
	password := "$2024Summer!"

	hashed, err := HashPassword(password)
	g.Expect(err).To(BeNil())
	g.Expect(hashed).NotTo(Equal(password))
	g.Expect(hashed).To(HavePrefix("bcrypt:"))
	g.Expect(len(hashed)).To(BeNumerically(">", 60))
}

// TestMatchPassword tests password matching functionality.
func TestMatchPassword(t *testing.T) {
	g := NewGomegaWithT(t)

	password := "MySecurePassword123!"

	// Hash the password
	hashed, err := HashPassword(password)
	g.Expect(err).To(BeNil())

	// Correct password should match
	matched, err := MatchPassword(password, hashed)
	g.Expect(err).To(BeNil())
	g.Expect(matched).To(BeTrue())

	// Wrong password should not match
	matched, err = MatchPassword("WrongPassword", hashed)
	g.Expect(err).To(BeNil())
	g.Expect(matched).To(BeFalse())
}

// TestMatchPasswordInvalidHash tests matching against invalid hash.
func TestMatchPasswordInvalidHash(t *testing.T) {
	g := NewGomegaWithT(t)

	password := "MySecurePassword123!"
	invalidHash := "not-a-bcrypt-hash"

	matched, err := MatchPassword(password, invalidHash)
	g.Expect(err).NotTo(BeNil())
	g.Expect(matched).To(BeFalse())
}

// TestPasswordHashAndMatch tests complete workflow.
func TestPasswordHashAndMatch(t *testing.T) {
	g := NewGomegaWithT(t)

	testCases := []struct {
		password string
		match    string
		expected bool
	}{
		{"SimplePassword", "SimplePassword", true},
		{"SimplePassword", "WrongPassword", false},
		{"C0mpl3x!P@ssw0rd", "C0mpl3x!P@ssw0rd", true},
		{"C0mpl3x!P@ssw0rd", "c0mpl3x!p@ssw0rd", false},
		{"short", "short", true},
		{"VeryLongPasswordWithLotsOfCharacters123!@#", "VeryLongPasswordWithLotsOfCharacters123!@#", true},
	}

	for _, tc := range testCases {
		hashed, err := HashPassword(tc.password)
		g.Expect(err).To(BeNil())

		matched, err := MatchPassword(tc.match, hashed)
		g.Expect(err).To(BeNil())
		g.Expect(matched).To(Equal(tc.expected))
	}
}
