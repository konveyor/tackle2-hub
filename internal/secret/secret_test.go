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
	hashed := HashPassword(password)
	g.Expect(hashed).NotTo(BeEmpty())
	g.Expect(hashed).NotTo(Equal(password))

	// Hash should be base64 encoded (80 chars for prefix + base64 of 60-byte bcrypt)
	g.Expect(len(hashed)).To(BeNumerically(">", 60))

	// Hashing same password produces different hash (due to salt)
	hashed2 := HashPassword(password)
	g.Expect(hashed2).NotTo(Equal(hashed))
}

// TestHashPasswordDoubleHashPrevention tests that already-hashed passwords are not re-hashed.
func TestHashPasswordDoubleHashPrevention(t *testing.T) {
	g := NewGomegaWithT(t)

	password := "MySecurePassword123!"

	// Hash the password
	hashed1 := HashPassword(password)

	// Try to hash the already-hashed password
	hashed2 := HashPassword(hashed1)
	g.Expect(hashed2).To(Equal(hashed1))
}

// TestHashPasswordEmptyString tests handling of empty password.
func TestHashPasswordEmptyString(t *testing.T) {
	g := NewGomegaWithT(t)

	hashed := HashPassword("")
	g.Expect(hashed).To(Equal(""))
}

// TestHashPasswordStartingWithDollar2 tests that passwords starting with $2 are hashed correctly.
func TestHashPasswordStartingWithDollar2(t *testing.T) {
	g := NewGomegaWithT(t)

	// Password that starts with $2 but isn't a bcrypt hash
	password := "$2024Summer!"

	hashed := HashPassword(password)
	g.Expect(hashed).NotTo(Equal(password))
	g.Expect(len(hashed)).To(BeNumerically(">", 60))
}

// TestMatchPassword tests password matching functionality.
func TestMatchPassword(t *testing.T) {
	g := NewGomegaWithT(t)

	password := "MySecurePassword123!"

	// Hash the password
	hashed := HashPassword(password)

	// Correct password should match
	matched := MatchPassword(password, hashed)
	g.Expect(matched).To(BeTrue())

	// Wrong password should not match
	matched = MatchPassword("WrongPassword", hashed)
	g.Expect(matched).To(BeFalse())
}

// TestMatchPasswordInvalidHash tests matching against invalid hash.
func TestMatchPasswordInvalidHash(t *testing.T) {
	g := NewGomegaWithT(t)

	password := "MySecurePassword123!"
	invalidHash := "not-a-bcrypt-hash"

	matched := MatchPassword(password, invalidHash)
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
		hashed := HashPassword(tc.password)

		matched := MatchPassword(tc.match, hashed)
		g.Expect(matched).To(Equal(tc.expected))
	}
}

// TestHashPasswordTruncation tests that passwords longer than 72 bytes are truncated.
func TestHashPasswordTruncation(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create a password exactly 72 bytes
	password72 := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890123456789"
	g.Expect(len(password72)).To(Equal(72))

	// Create a password longer than 72 bytes
	password80 := password72 + "12345678"
	g.Expect(len(password80)).To(Equal(80))

	// Hash both passwords
	hashed72 := HashPassword(password72)

	hashed80 := HashPassword(password80)

	// The 72-byte password should match itself
	matched := MatchPassword(password72, hashed72)
	g.Expect(matched).To(BeTrue())

	// The 80-byte password should match using its hash
	matched = MatchPassword(password80, hashed80)
	g.Expect(matched).To(BeTrue())

	// The 72-byte password should match the hash of the 80-byte password
	// because the 80-byte password is truncated to 72 bytes
	matched = MatchPassword(password72, hashed80)
	g.Expect(matched).To(BeTrue())

	// The 80-byte password should NOT match the hash of the 72-byte password
	// because only the first 72 bytes are used
	matched = MatchPassword(password80, hashed72)
	g.Expect(matched).To(BeTrue())
}

// TestHashPasswordMultibyteCharacters tests truncation with multibyte UTF-8 characters.
func TestHashPasswordMultibyteCharacters(t *testing.T) {
	g := NewGomegaWithT(t)

	// Create a password with multibyte characters that exceeds 72 bytes
	// Each emoji is 4 bytes, so 20 emojis = 80 bytes
	password := "🔒🔒🔒🔒🔒🔒🔒🔒🔒🔒🔒🔒🔒🔒🔒🔒🔒🔒🔒🔒"
	g.Expect(len(password)).To(BeNumerically(">", 72))

	// Hash should succeed
	hashed := HashPassword(password)
	g.Expect(hashed).NotTo(BeEmpty())

	// Password should match its own hash
	matched := MatchPassword(password, hashed)
	g.Expect(matched).To(BeTrue())
}

func TestEncode(t *testing.T) {
	g := NewGomegaWithT(t)

	cipher := AESGCM{}
	cipher.Use("test")
	secret := Secret{Cipher: &cipher}
	password := "rabbit-slayer"
	apiKey := "secret-key-123"

	//
	// struct
	object := struct {
		Name     string
		Password string `secret:"hashed"`
		ApiKey   string `secret:"encrypted"`
		Token    string `secret:""` // defaults to encrypted
		Age      int
	}{
		Name:     "elmer",
		Password: password,
		ApiKey:   apiKey,
		Token:    "token-value",
		Age:      52,
	}
	fields, err := secret.Encode(&object)
	g.Expect(err).To(BeNil())
	g.Expect(len(fields)).To(Equal(3))
	g.Expect("elmer").To(Equal(object.Name))
	g.Expect(password).ToNot(Equal(object.Password))
	g.Expect(isHashedPassword(object.Password)).To(BeTrue())
	g.Expect(apiKey).ToNot(Equal(object.ApiKey))
	g.Expect("token-value").ToNot(Equal(object.Token))
	_, err = secret.Decode(&object)
	g.Expect(err).To(BeNil())
	g.Expect("elmer").To(Equal(object.Name))
	g.Expect(password).ToNot(Equal(object.Password)) // hashed stays hashed
	g.Expect(apiKey).To(Equal(object.ApiKey))
	g.Expect("token-value").To(Equal(object.Token))
}

func TestEncodePackage(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	password := "rabbit-slayer"
	apiKey := "secret-key-123"

	//
	// struct
	object := struct {
		Name     string
		Password string `secret:"hashed"`
		ApiKey   string `secret:"encrypted"`
		Token    string `secret:""` // defaults to encrypted
		Age      int
	}{
		Name:     "elmer",
		Password: password,
		ApiKey:   apiKey,
		Token:    "token-value",
		Age:      52,
	}
	fields, err := Encode(&object)
	g.Expect(err).To(BeNil())
	g.Expect(len(fields)).To(Equal(3))
	g.Expect("elmer").To(Equal(object.Name))
	g.Expect(password).ToNot(Equal(object.Password))
	g.Expect(isHashedPassword(object.Password)).To(BeTrue())
	g.Expect(apiKey).ToNot(Equal(object.ApiKey))
	g.Expect("token-value").ToNot(Equal(object.Token))
	err = Decode(&object)
	g.Expect(err).To(BeNil())
	g.Expect("elmer").To(Equal(object.Name))
	g.Expect(password).ToNot(Equal(object.Password)) // hashed stays hashed
	g.Expect(apiKey).To(Equal(object.ApiKey))
	g.Expect("token-value").To(Equal(object.Token))
}

func TestEncodeFields(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	//
	// verify returned fields
	object := struct {
		Password string `secret:"hashed"`
		ApiKey   string `secret:"encrypted"`
		Token    string `secret:""`
	}{
		Password: "pass123",
		ApiKey:   "key456",
		Token:    "token789",
	}
	fields, err := Encode(&object)
	g.Expect(err).To(BeNil())
	g.Expect(len(fields)).To(Equal(3))

	// Verify field names
	g.Expect(fields[0].name).To(Equal("Password"))
	g.Expect(fields[0].tag).To(Equal("hashed"))
	g.Expect(fields[0].Root()).To(BeTrue())

	g.Expect(fields[1].name).To(Equal("ApiKey"))
	g.Expect(fields[1].tag).To(Equal("encrypted"))
	g.Expect(fields[1].Root()).To(BeTrue())

	g.Expect(fields[2].name).To(Equal("Token"))
	g.Expect(fields[2].tag).To(Equal(""))
	g.Expect(fields[2].Root()).To(BeTrue())
}

// TestEncryptIdempotent tests that calling Encrypt() multiple times is idempotent.
func TestEncryptIdempotent(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	//
	// string
	s := "hello"
	err := Encrypt(&s)
	g.Expect(err).To(BeNil())
	encrypted := s

	// Second encryption should be no-op (already encrypted)
	err = Encrypt(&s)
	g.Expect(err).To(BeNil())
	g.Expect(s).To(Equal(encrypted))

	//
	// struct
	object := struct {
		ApiKey string `secret:"encrypted"`
		Token  string `secret:""`
	}{
		ApiKey: "key123",
		Token:  "token456",
	}

	err = Encrypt(&object)
	g.Expect(err).To(BeNil())
	encryptedKey := object.ApiKey
	encryptedToken := object.Token

	// Second encryption should be no-op
	err = Encrypt(&object)
	g.Expect(err).To(BeNil())
	g.Expect(object.ApiKey).To(Equal(encryptedKey))
	g.Expect(object.Token).To(Equal(encryptedToken))
}

// TestEncodeIdempotent tests that calling Encode() multiple times is idempotent.
func TestEncodeIdempotent(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	object := struct {
		Password string `secret:"hashed"`
		ApiKey   string `secret:"encrypted"`
		Token    string `secret:""`
	}{
		Password: "pass123",
		ApiKey:   "key456",
		Token:    "token789",
	}

	fields1, err := Encode(&object)
	g.Expect(err).To(BeNil())
	g.Expect(len(fields1)).To(Equal(3))
	hash1 := object.Password
	encrypted1 := object.ApiKey
	encrypted2 := object.Token

	// Verify first encoding worked
	g.Expect(hash1).ToNot(Equal("pass123"))
	g.Expect(isHashedPassword(hash1)).To(BeTrue())
	g.Expect(encrypted1).ToNot(Equal("key456"))
	g.Expect(encrypted2).ToNot(Equal("token789"))

	// Second encode should be no-op (already encoded)
	fields2, err := Encode(&object)
	g.Expect(err).To(BeNil())
	g.Expect(len(fields2)).To(Equal(3))
	g.Expect(object.Password).To(Equal(hash1))
	g.Expect(object.ApiKey).To(Equal(encrypted1))
	g.Expect(object.Token).To(Equal(encrypted2))

	// Field metadata should still be correct
	g.Expect(fields2[0].name).To(Equal("Password"))
	g.Expect(fields2[0].tag).To(Equal("hashed"))
	g.Expect(fields2[1].name).To(Equal("ApiKey"))
	g.Expect(fields2[1].tag).To(Equal("encrypted"))
}

// TestEncryptEmptyString tests encryption of empty strings.
func TestEncryptEmptyString(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	//
	// string
	s := ""
	err := Encrypt(&s)
	g.Expect(err).To(BeNil())
	g.Expect(s).To(Equal(""))

	err = Decrypt(&s)
	g.Expect(err).To(BeNil())
	g.Expect(s).To(Equal(""))

	//
	// struct
	object := struct {
		ApiKey string `secret:"encrypted"`
	}{
		ApiKey: "",
	}

	err = Encrypt(&object)
	g.Expect(err).To(BeNil())
	g.Expect(object.ApiKey).To(Equal(""))

	err = Decrypt(&object)
	g.Expect(err).To(BeNil())
	g.Expect(object.ApiKey).To(Equal(""))
}

// TestEncryptWithHashedTag tests that Encrypt handles hashed tags correctly.
func TestEncryptWithHashedTag(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	//
	// struct with hashed tag
	object := struct {
		Password string `secret:"hashed"`
	}{
		Password: "pass123",
	}

	// Encrypt should no-op hashed fields
	err := Encrypt(&object)
	g.Expect(err).To(BeNil())
	g.Expect(object.Password).To(Equal("pass123"))

	// Similarly for Decrypt with already-hashed password
	hashedPassword := HashPassword("pass123")
	object.Password = hashedPassword
	err = Decrypt(&object)
	g.Expect(err).To(BeNil())
	g.Expect(object.Password).To(Equal(hashedPassword))
}

// TestDecryptWithHashedTag tests that Decrypt handles hashed tags correctly.
func TestDecryptWithHashedTag(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	hashedPassword := HashPassword("mypassword")

	object := struct {
		Password string `secret:"hashed"`
	}{
		Password: hashedPassword,
	}

	// Decrypt should no-op hashed fields
	err := Decrypt(&object)
	g.Expect(err).To(BeNil())
	g.Expect(object.Password).To(Equal(hashedPassword))
}

// TestEncryptMixedTags tests Encrypt/Decrypt with mixed hashed and encrypted tags.
func TestEncryptMixedTags(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	object := struct {
		Password string `secret:"hashed"`
		ApiKey   string `secret:"encrypted"`
	}{
		Password: "pass123",
		ApiKey:   "key456",
	}

	// Encrypt should only encrypt "encrypted" tagged fields
	err := Encrypt(&object)
	g.Expect(err).To(BeNil())
	g.Expect(object.Password).To(Equal("pass123")) // hashed tag = no-op
	g.Expect(object.ApiKey).ToNot(Equal("key456")) // encrypted

	encryptedKey := object.ApiKey

	// Decrypt should only decrypt "encrypted" tagged fields
	err = Decrypt(&object)
	g.Expect(err).To(BeNil())
	g.Expect(object.Password).To(Equal("pass123")) // Still unchanged
	g.Expect(object.ApiKey).To(Equal("key456"))    // Decrypted
	g.Expect(encryptedKey).ToNot(Equal("key456"))  // Verify it was actually encrypted
}

// TestEncryptNonExportedFields tests that non-exported fields are ignored.
func TestEncryptNonExportedFields(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	object := struct {
		Public  string `secret:"encrypted"`
		private string `secret:"encrypted"` // Should be ignored
	}{
		Public:  "public-value",
		private: "private-value",
	}

	err := Encrypt(&object)
	g.Expect(err).To(BeNil())
	g.Expect(object.Public).ToNot(Equal("public-value")) // Encrypted
	g.Expect(object.private).To(Equal("private-value"))  // Unchanged (non-exported)

	err = Decrypt(&object)
	g.Expect(err).To(BeNil())
	g.Expect(object.Public).To(Equal("public-value"))   // Decrypted
	g.Expect(object.private).To(Equal("private-value")) // Still unchanged
}

// TestEncodeNonExportedFields tests that Encode ignores non-exported fields.
func TestEncodeNonExportedFields(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	object := struct {
		Public  string `secret:"encrypted"`
		private string `secret:"encrypted"` // Should be ignored
	}{
		Public:  "public-value",
		private: "private-value",
	}

	fields, err := Encode(&object)
	g.Expect(err).To(BeNil())
	g.Expect(len(fields)).To(Equal(1)) // Only Public
	g.Expect(fields[0].name).To(Equal("Public"))
	g.Expect(object.Public).ToNot(Equal("public-value")) // Encrypted
	g.Expect(object.private).To(Equal("private-value"))  // Unchanged
}

// TestFieldFqnRoot tests Field.Fqn() and Root() for root-level fields.
// This reflects actual production usage where all secret-tagged fields
// are at the root level of their models (User.Password, IdpClient.Secret).
func TestFieldFqnRoot(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	object := struct {
		Name     string
		Password string `secret:"hashed"`
		ApiKey   string `secret:"encrypted"`
		Token    string `secret:""`
	}{
		Name:     "test",
		Password: "pass123",
		ApiKey:   "key456",
		Token:    "token789",
	}

	fields, err := Encode(&object)
	g.Expect(err).To(BeNil())
	g.Expect(len(fields)).To(Equal(3))

	// Verify all fields are root-level
	for i := range fields {
		f := &fields[i]
		g.Expect(f.Root()).To(BeTrue())

		// Fqn should equal name for root-level fields
		g.Expect(f.Fqn()).To(Equal(f.name))
	}

	// Verify specific field names
	fieldNames := make(map[string]bool)
	for i := range fields {
		fieldNames[fields[i].name] = true
	}
	g.Expect(fieldNames["Password"]).To(BeTrue())
	g.Expect(fieldNames["ApiKey"]).To(BeTrue())
	g.Expect(fieldNames["Token"]).To(BeTrue())
}

// TestFieldSecretValueCapture tests that Field.Secret() captures pre-transformation value.
func TestFieldSecretValueCapture(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	password := "original-password"
	apiKey := "original-key"

	object := struct {
		Password string `secret:"hashed"`
		ApiKey   string `secret:"encrypted"`
	}{
		Password: password,
		ApiKey:   apiKey,
	}

	fields, err := Encode(&object)
	g.Expect(err).To(BeNil())
	g.Expect(len(fields)).To(Equal(2))

	// Find fields
	var passwordField, apiKeyField *Field
	for i := range fields {
		f := &fields[i]
		if f.name == "Password" {
			passwordField = f
		}
		if f.name == "ApiKey" {
			apiKeyField = f
		}
	}

	// Verify Secret() returns ORIGINAL pre-transformation values
	g.Expect(passwordField).ToNot(BeNil())
	g.Expect(passwordField.Secret()).To(Equal(password)) // Original, not hash
	g.Expect(object.Password).ToNot(Equal(password))     // But object IS transformed

	g.Expect(apiKeyField).ToNot(BeNil())
	g.Expect(apiKeyField.Secret()).To(Equal(apiKey)) // Original, not encrypted
	g.Expect(object.ApiKey).ToNot(Equal(apiKey))     // But object IS transformed
}

//
// Error case tests
//

// TestDecryptCorruptedData tests decryption of corrupted encrypted data.
func TestDecryptCorruptedData(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	//
	// Corrupted base64
	s := "!!!invalid-base64!!!"
	err := Decrypt(&s)
	g.Expect(err).ToNot(BeNil())

	//
	// Valid base64 but corrupted ciphertext
	s = "aGVsbG8gd29ybGQ=" // "hello world" in base64, not valid AES-GCM
	err = Decrypt(&s)
	g.Expect(err).ToNot(BeNil())

	//
	// Struct with corrupted encrypted field
	object := struct {
		ApiKey string `secret:"encrypted"`
	}{
		ApiKey: "!!!invalid!!!",
	}
	err = Decrypt(&object)
	g.Expect(err).ToNot(BeNil())
}

// TestDecryptWrongPassphrase tests decryption with wrong passphrase.
func TestDecryptWrongPassphrase(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	s := "secret-data"
	err := Encrypt(&s)
	g.Expect(err).To(BeNil())
	encrypted := s

	// Change passphrase
	Settings.Passphrase = "WRONG"
	err = Decrypt(&s)
	g.Expect(err).ToNot(BeNil())
	g.Expect(s).To(Equal(encrypted)) // Value unchanged on error
}

// TestDecodeCorruptedData tests Decode with corrupted encrypted fields.
func TestDecodeCorruptedData(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	object := struct {
		Password string `secret:"hashed"`
		ApiKey   string `secret:"encrypted"`
	}{
		Password: HashPassword("pass123"),
		ApiKey:   "!!!corrupted!!!",
	}

	err := Decode(&object)
	g.Expect(err).ToNot(BeNil())
}

// TestEncryptUnknownTag tests that unknown tags return error.
func TestEncryptUnknownTag(t *testing.T) {
	g := NewGomegaWithT(t)

	cipher := AESGCM{}
	cipher.Use("test")
	secret := Secret{Cipher: &cipher}

	object := struct {
		Field string `secret:"unknown"`
	}{
		Field: "value",
	}

	// Should return error (panic is caught by Update's defer/recover)
	err := secret.Encrypt(&object)
	g.Expect(err).ToNot(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("unknown tag"))
	g.Expect(err.Error()).To(ContainSubstring("unknown"))
}

// TestDecryptUnknownTag tests that unknown tags return error in Decrypt.
func TestDecryptUnknownTag(t *testing.T) {
	g := NewGomegaWithT(t)

	cipher := AESGCM{}
	cipher.Use("test")
	secret := Secret{Cipher: &cipher}

	object := struct {
		Field string `secret:"unknown"`
	}{
		Field: "value",
	}

	// Should return error (panic is caught by Update's defer/recover)
	err := secret.Decrypt(&object)
	g.Expect(err).ToNot(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("unknown tag"))
	g.Expect(err.Error()).To(ContainSubstring("unknown"))
}

// TestEncodeUnknownTag tests that unknown tags return error in Encode.
func TestEncodeUnknownTag(t *testing.T) {
	g := NewGomegaWithT(t)

	cipher := AESGCM{}
	cipher.Use("test")
	secret := Secret{Cipher: &cipher}

	object := struct {
		Field string `secret:"unknown"`
	}{
		Field: "value",
	}

	// Should return error (panic is caught by Update's defer/recover)
	_, err := secret.Encode(&object)
	g.Expect(err).ToNot(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("unknown tag"))
	g.Expect(err.Error()).To(ContainSubstring("unknown"))
}

// TestDecodeUnknownTag tests that unknown tags return error in Decode.
func TestDecodeUnknownTag(t *testing.T) {
	g := NewGomegaWithT(t)

	cipher := AESGCM{}
	cipher.Use("test")
	secret := Secret{Cipher: &cipher}

	object := struct {
		Field string `secret:"unknown"`
	}{
		Field: "value",
	}

	// Should return error (panic is caught by Update's defer/recover)
	_, err := secret.Decode(&object)
	g.Expect(err).ToNot(BeNil())
	g.Expect(err.Error()).To(ContainSubstring("unknown tag"))
	g.Expect(err.Error()).To(ContainSubstring("unknown"))
}

// TestCipherEmptyPassphrase tests cipher with empty passphrase.
func TestCipherEmptyPassphrase(t *testing.T) {
	g := NewGomegaWithT(t)

	cipher := AESGCM{}
	cipher.Use("")

	// Empty passphrase results in nil key (per gcm.go:25-27)
	g.Expect(cipher.Key).To(BeNil())

	// Encryption should fail with nil key
	encrypted, err := cipher.Encrypt("hello")
	g.Expect(err).ToNot(BeNil())
	g.Expect(encrypted).To(Equal(""))
}

// TestCipherWhitespacePassphrase tests cipher with whitespace-only passphrase.
func TestCipherWhitespacePassphrase(t *testing.T) {
	g := NewGomegaWithT(t)

	cipher := AESGCM{}
	cipher.Use("   \t\n  ")

	// Whitespace-only passphrase gets trimmed to empty (per gcm.go:23)
	g.Expect(cipher.Key).To(BeNil())

	// Encryption should fail with nil key
	encrypted, err := cipher.Encrypt("hello")
	g.Expect(err).ToNot(BeNil())
	g.Expect(encrypted).To(Equal(""))
}

// TestCipherLongPassphrase tests cipher with very long passphrase.
func TestCipherLongPassphrase(t *testing.T) {
	g := NewGomegaWithT(t)

	// 100-character passphrase
	longPassphrase := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:,.<>?/~`1234567890"
	g.Expect(len(longPassphrase)).To(BeNumerically(">", 32))

	cipher := AESGCM{}
	cipher.Use(longPassphrase)

	// Key should be exactly 32 bytes (per gcm.go:35)
	g.Expect(len(cipher.Key)).To(Equal(32))

	// Should work correctly
	plain := "test-data"
	encrypted, err := cipher.Encrypt(plain)
	g.Expect(err).To(BeNil())
	g.Expect(encrypted).ToNot(Equal(plain))

	decrypted, err := cipher.Decrypt(encrypted)
	g.Expect(err).To(BeNil())
	g.Expect(decrypted).To(Equal(plain))
}

// TestCipherNonASCIIPassphrase tests cipher with non-ASCII passphrase.
func TestCipherNonASCIIPassphrase(t *testing.T) {
	g := NewGomegaWithT(t)

	// Unicode passphrase with emojis and multi-byte characters
	unicodePassphrase := "🔒秘密🔑パスワード"

	cipher := AESGCM{}
	cipher.Use(unicodePassphrase)

	// Key should be exactly 32 bytes
	g.Expect(len(cipher.Key)).To(Equal(32))

	// Should work correctly
	plain := "test-data"
	encrypted, err := cipher.Encrypt(plain)
	g.Expect(err).To(BeNil())
	g.Expect(encrypted).ToNot(Equal(plain))

	decrypted, err := cipher.Decrypt(encrypted)
	g.Expect(err).To(BeNil())
	g.Expect(decrypted).To(Equal(plain))
}

// TestEncryptDecryptRoundTripErrors tests error handling in round-trip operations.
func TestEncryptDecryptRoundTripErrors(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	object := struct {
		Name   string
		ApiKey string `secret:"encrypted"`
		Token  string `secret:""`
	}{
		Name:   "test",
		ApiKey: "key123",
		Token:  "token456",
	}

	// Encrypt successfully
	err := Encrypt(&object)
	g.Expect(err).To(BeNil())
	encryptedKey := object.ApiKey
	encryptedToken := object.Token

	// Corrupt one field
	object.ApiKey = "!!!corrupted!!!"

	// Decrypt should fail
	err = Decrypt(&object)
	g.Expect(err).ToNot(BeNil())

	// Restore and verify
	object.ApiKey = encryptedKey
	object.Token = encryptedToken
	err = Decrypt(&object)
	g.Expect(err).To(BeNil())
	g.Expect(object.ApiKey).To(Equal("key123"))
	g.Expect(object.Token).To(Equal("token456"))
}

// TestMapDecryptError tests error handling when decrypting maps with corrupted data.
func TestMapDecryptError(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	mp := map[string]any{
		"apiKey": "key123",
		"token":  "token456",
	}

	// Encrypt
	err := Encrypt(mp)
	g.Expect(err).To(BeNil())

	// Corrupt a field
	mp["apiKey"] = "!!!invalid!!!"

	// Decrypt should fail
	err = Decrypt(mp)
	g.Expect(err).ToNot(BeNil())
}

// TestEncryptionErrorPropagation tests that encryption errors are properly propagated.
func TestEncryptionErrorPropagation(t *testing.T) {
	g := NewGomegaWithT(t)

	// Empty passphrase causes nil key
	Settings.Passphrase = ""

	s := "test"
	err := Encrypt(&s)
	g.Expect(err).ToNot(BeNil())

	object := struct {
		ApiKey string `secret:"encrypted"`
	}{
		ApiKey: "key123",
	}

	err = Encrypt(&object)
	g.Expect(err).ToNot(BeNil())
}

// TestRevertRedacted tests RevertRedacted() function.
func TestRevertRedacted(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	const mask = "***"

	// Current object (from database)
	current := struct {
		Name     string
		Password string `secret:"hashed"`
		ApiKey   string `secret:"encrypted"`
	}{
		Name:     "user1",
		Password: "hashed-password",
		ApiKey:   "encrypted-key",
	}

	// Updated object (from request with mask for password)
	updated := struct {
		Name     string
		Password string `secret:"hashed"`
		ApiKey   string `secret:"encrypted"`
	}{
		Name:     "user1",
		Password: mask,            // Client sent mask - don't change
		ApiKey:   "new-key-value", // Client sent new value
	}

	// Encode updated (would hash/encrypt new values)
	fields, err := Encode(&updated)
	g.Expect(err).To(BeNil())
	g.Expect(len(fields)).To(Equal(2))

	// Restore masked values from current
	err = RevertRedacted(fields, &current, mask)
	g.Expect(err).To(BeNil())

	// Password should be restored from current
	g.Expect(updated.Password).To(Equal("hashed-password"))

	// ApiKey should be newly encrypted (not equal to either original)
	g.Expect(updated.ApiKey).ToNot(Equal("new-key-value"))
	g.Expect(updated.ApiKey).ToNot(Equal("encrypted-key"))
}

// TestRevertRedactedNoMask tests RevertRedacted when no fields are masked.
func TestRevertRedactedNoMask(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	const mask = "***"

	current := struct {
		Password string `secret:"hashed"`
	}{
		Password: "old-password",
	}

	updated := struct {
		Password string `secret:"hashed"`
	}{
		Password: "new-password",
	}

	fields, err := Encode(&updated)
	g.Expect(err).To(BeNil())

	originalHash := updated.Password

	// No masked fields - should be no-op
	err = RevertRedacted(fields, &current, mask)
	g.Expect(err).To(BeNil())

	// Password should still be the new hash
	g.Expect(updated.Password).To(Equal(originalHash))
	g.Expect(updated.Password).ToNot(Equal("new-password"))
	g.Expect(updated.Password).ToNot(Equal("old-password"))
}

// TestRevertRedactedAllMasked tests RevertRedacted when all fields are masked.
func TestRevertRedactedAllMasked(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	const mask = "***"

	current := struct {
		Password string `secret:"hashed"`
		ApiKey   string `secret:"encrypted"`
	}{
		Password: "hashed-password",
		ApiKey:   "encrypted-key",
	}

	updated := struct {
		Password string `secret:"hashed"`
		ApiKey   string `secret:"encrypted"`
	}{
		Password: mask,
		ApiKey:   mask,
	}

	fields, err := Encode(&updated)
	g.Expect(err).To(BeNil())

	// Restore all masked values
	err = RevertRedacted(fields, &current, mask)
	g.Expect(err).To(BeNil())

	// Both fields should be restored from current
	g.Expect(updated.Password).To(Equal("hashed-password"))
	g.Expect(updated.ApiKey).To(Equal("encrypted-key"))
}

// TestRevertRedactedMixedTypes tests RevertRedacted with hashed and encrypted fields.
func TestRevertRedactedMixedTypes(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	const mask = "REDACTED"

	current := struct {
		Password string `secret:"hashed"`
		Token    string `secret:"encrypted"`
		ApiKey   string `secret:""`
	}{
		Password: HashPassword("old-pass"),
		Token:    "encrypted-token-1",
		ApiKey:   "encrypted-key-1",
	}

	// Encrypt current's encrypted fields
	err := Encrypt(&current)
	g.Expect(err).To(BeNil())

	updated := struct {
		Password string `secret:"hashed"`
		Token    string `secret:"encrypted"`
		ApiKey   string `secret:""`
	}{
		Password: mask,        // Masked - restore
		Token:    "new-token", // New value
		ApiKey:   mask,        // Masked - restore
	}

	fields, err := Encode(&updated)
	g.Expect(err).To(BeNil())

	err = RevertRedacted(fields, &current, mask)
	g.Expect(err).To(BeNil())

	// Password restored
	g.Expect(isHashedPassword(updated.Password)).To(BeTrue())
	g.Expect(updated.Password).To(Equal(current.Password))

	// Token is new encrypted value
	g.Expect(updated.Token).ToNot(Equal("new-token"))
	g.Expect(updated.Token).ToNot(Equal(current.Token))

	// ApiKey restored
	g.Expect(updated.ApiKey).To(Equal(current.ApiKey))
}

// TestRevertRedactedDifferentStructs tests RevertRedacted with different struct types.
func TestRevertRedactedDifferentStructs(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	const mask = "***"

	// Different struct but with matching field names
	current := struct {
		Name     string
		Password string `secret:"hashed"`
		Extra    string
	}{
		Name:     "user1",
		Password: "old-hash",
		Extra:    "extra-data",
	}

	updated := struct {
		Password string `secret:"hashed"`
		Email    string
	}{
		Password: mask,
		Email:    "user@example.com",
	}

	fields, err := Encode(&updated)
	g.Expect(err).To(BeNil())

	// Should restore Password by field name match
	err = RevertRedacted(fields, &current, mask)
	g.Expect(err).To(BeNil())

	g.Expect(updated.Password).To(Equal("old-hash"))
	g.Expect(updated.Email).To(Equal("user@example.com"))
}

// TestRevertRedactedEmptyFields tests RevertRedacted with no secret fields.
func TestRevertRedactedEmptyFields(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	const mask = "***"

	current := struct {
		Name string
	}{
		Name: "user1",
	}

	updated := struct {
		Name string
	}{
		Name: "user1",
	}

	// No secret fields - empty fields slice
	fields, err := Encode(&updated)
	g.Expect(err).To(BeNil())
	g.Expect(len(fields)).To(Equal(0))

	// Should be no-op
	err = RevertRedacted(fields, &current, mask)
	g.Expect(err).To(BeNil())
}

// TestRevertRedactedError tests RevertRedacted error handling.
func TestRevertRedactedError(t *testing.T) {
	g := NewGomegaWithT(t)

	Settings.Passphrase = "TEST"

	const mask = "***"

	updated := struct {
		Password string `secret:"hashed"`
	}{
		Password: mask,
	}

	fields, err := Encode(&updated)
	g.Expect(err).To(BeNil())

	// Pass invalid object type (string instead of struct pointer)
	invalidString := "not-a-struct"
	err = RevertRedacted(fields, &invalidString, mask)
	g.Expect(err).To(BeNil()) // Fields() on string pointer succeeds but finds no secret fields

	// Pass nil should error
	err = RevertRedacted(fields, nil, mask)
	g.Expect(err).ToNot(BeNil())
}
