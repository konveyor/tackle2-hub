package ptr

import (
	"testing"

	"github.com/onsi/gomega"
)

// TestCopyNil verifies Copy handles nil input.
func TestCopyNil(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	var p *string
	result := Copy(p)

	g.Expect(result).To(gomega.BeNil())
}

// TestCopyBasicTypes verifies Copy handles basic types.
func TestCopyBasicTypes(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	type Basic struct {
		String string
		Int    int
		Uint   uint
		Bool   bool
		Float  float64
	}

	original := &Basic{
		String: "test",
		Int:    42,
		Uint:   100,
		Bool:   true,
		Float:  3.14,
	}

	copied := Copy(original)

	g.Expect(copied).NotTo(gomega.BeIdenticalTo(original))
	g.Expect(copied).To(gomega.Equal(original))

	// Verify independence
	copied.String = "modified"
	g.Expect(original.String).To(gomega.Equal("test"))
}

// TestCopySlice verifies Copy deep copies slices.
func TestCopySlice(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	type WithSlice struct {
		Items []string
	}

	original := &WithSlice{
		Items: []string{"one", "two", "three"},
	}

	copied := Copy(original)

	g.Expect(copied).To(gomega.Equal(original))
	g.Expect(copied.Items).NotTo(gomega.BeIdenticalTo(original.Items))

	// Verify independence - modifying copy doesn't affect original
	copied.Items[0] = "MODIFIED"
	g.Expect(original.Items[0]).To(gomega.Equal("one"))
	g.Expect(copied.Items[0]).To(gomega.Equal("MODIFIED"))
}

// TestCopySliceOfStructs verifies Copy handles slices of structs (critical for User/IdpClient).
func TestCopySliceOfStructs(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	type Role struct {
		ID   uint
		Name string
	}

	type User struct {
		Login string
		Roles []Role
	}

	original := &User{
		Login: "testuser",
		Roles: []Role{
			{ID: 1, Name: "admin"},
			{ID: 2, Name: "user"},
		},
	}

	copied := Copy(original)

	g.Expect(copied).To(gomega.Equal(original))
	g.Expect(copied.Roles).NotTo(gomega.BeIdenticalTo(original.Roles))

	// Critical test: modifying nested struct in copy doesn't affect original
	copied.Roles[0].Name = "MODIFIED"
	g.Expect(original.Roles[0].Name).To(gomega.Equal("admin"))
	g.Expect(copied.Roles[0].Name).To(gomega.Equal("MODIFIED"))
}

// TestCopyMultipleSlices verifies Copy handles structs with multiple slices (like IdpClient).
func TestCopyMultipleSlices(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	type Client struct {
		Secret       string
		Grants       []string
		RedirectURIs []string
		Scopes       []string
	}

	original := &Client{
		Secret:       "secret123",
		Grants:       []string{"authorization_code", "refresh_token"},
		RedirectURIs: []string{"http://localhost/callback"},
		Scopes:       []string{"openid", "profile"},
	}

	copied := Copy(original)

	g.Expect(copied).To(gomega.Equal(original))

	// Verify all slices are independent
	copied.Grants[0] = "MODIFIED"
	copied.RedirectURIs[0] = "MODIFIED"
	copied.Scopes[0] = "MODIFIED"

	g.Expect(original.Grants[0]).To(gomega.Equal("authorization_code"))
	g.Expect(original.RedirectURIs[0]).To(gomega.Equal("http://localhost/callback"))
	g.Expect(original.Scopes[0]).To(gomega.Equal("openid"))
}

// TestCopyNilSlice verifies Copy handles nil slices correctly.
func TestCopyNilSlice(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	type WithNilSlice struct {
		Items []string
	}

	original := &WithNilSlice{
		Items: nil,
	}

	copied := Copy(original)

	g.Expect(copied.Items).To(gomega.BeNil())
}

// TestCopyEmptySlice verifies Copy preserves empty slices (different from nil).
func TestCopyEmptySlice(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	type WithEmptySlice struct {
		Items []string
	}

	original := &WithEmptySlice{
		Items: []string{},
	}

	copied := Copy(original)

	g.Expect(copied.Items).NotTo(gomega.BeNil())
	g.Expect(copied.Items).To(gomega.HaveLen(0))
}

// TestCopyPointerFields verifies Copy handles pointer fields.
func TestCopyPointerFields(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	type Inner struct {
		Value string
	}

	type WithPointer struct {
		Ptr *Inner
	}

	inner := &Inner{Value: "test"}
	original := &WithPointer{
		Ptr: inner,
	}

	copied := Copy(original)

	g.Expect(copied.Ptr).NotTo(gomega.BeIdenticalTo(original.Ptr))
	g.Expect(copied.Ptr).To(gomega.Equal(original.Ptr))

	// Verify independence
	copied.Ptr.Value = "MODIFIED"
	g.Expect(original.Ptr.Value).To(gomega.Equal("test"))
}

// TestCopyNilPointerField verifies Copy handles nil pointer fields.
func TestCopyNilPointerField(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	type Inner struct {
		Value string
	}

	type WithPointer struct {
		Ptr *Inner
	}

	original := &WithPointer{
		Ptr: nil,
	}

	copied := Copy(original)

	g.Expect(copied.Ptr).To(gomega.BeNil())
}

// TestCopyNestedPointers verifies Copy handles nested pointers.
func TestCopyNestedPointers(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	type Nested struct {
		Value string
	}

	type WithNestedPtr struct {
		PtrPtr **Nested
	}

	value := "test"
	nested := &Nested{Value: value}
	nestedPtr := &nested

	original := &WithNestedPtr{
		PtrPtr: nestedPtr,
	}

	copied := Copy(original)

	g.Expect(copied.PtrPtr).NotTo(gomega.BeIdenticalTo(original.PtrPtr))
	g.Expect(*copied.PtrPtr).NotTo(gomega.BeIdenticalTo(*original.PtrPtr))
	g.Expect(**copied.PtrPtr).To(gomega.Equal(**original.PtrPtr))

	// Verify independence
	(**copied.PtrPtr).Value = "MODIFIED"
	g.Expect((**original.PtrPtr).Value).To(gomega.Equal("test"))
}

// TestCopyMap verifies Copy deep copies maps.
func TestCopyMap(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	type WithMap struct {
		Data map[string]string
	}

	original := &WithMap{
		Data: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	copied := Copy(original)

	g.Expect(copied.Data).NotTo(gomega.BeIdenticalTo(original.Data))
	g.Expect(copied.Data).To(gomega.Equal(original.Data))

	// Verify independence
	copied.Data["key1"] = "MODIFIED"
	g.Expect(original.Data["key1"]).To(gomega.Equal("value1"))
}

// TestCopyMapWithStructValues verifies Copy deep copies map values.
func TestCopyMapWithStructValues(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	type Value struct {
		Field string
	}

	type WithMapOfStructs struct {
		Data map[string]Value
	}

	original := &WithMapOfStructs{
		Data: map[string]Value{
			"key1": {Field: "value1"},
			"key2": {Field: "value2"},
		},
	}

	copied := Copy(original)

	g.Expect(copied.Data).NotTo(gomega.BeIdenticalTo(original.Data))
	g.Expect(copied.Data).To(gomega.Equal(original.Data))

	// Verify independence
	val := copied.Data["key1"]
	val.Field = "MODIFIED"
	copied.Data["key1"] = val

	g.Expect(original.Data["key1"].Field).To(gomega.Equal("value1"))
}

// TestCopyNilMap verifies Copy handles nil maps.
func TestCopyNilMap(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	type WithNilMap struct {
		Data map[string]string
	}

	original := &WithNilMap{
		Data: nil,
	}

	copied := Copy(original)

	g.Expect(copied.Data).To(gomega.BeNil())
}

// TestCopyArray verifies Copy handles arrays.
func TestCopyArray(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	type WithArray struct {
		Items [3]string
	}

	original := &WithArray{
		Items: [3]string{"one", "two", "three"},
	}

	copied := Copy(original)

	g.Expect(copied).To(gomega.Equal(original))

	// Verify independence
	copied.Items[0] = "MODIFIED"
	g.Expect(original.Items[0]).To(gomega.Equal("one"))
}

// TestCopyComplex verifies Copy handles complex nested structures.
func TestCopyComplex(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	type Token struct {
		ID      uint
		Subject string
	}

	type Role struct {
		ID          uint
		Name        string
		Permissions []string
	}

	type User struct {
		ID       uint
		Login    string
		Password string
		Email    string
		Roles    []Role
		Tokens   []Token
		Metadata map[string]string
	}

	original := &User{
		ID:       1,
		Login:    "testuser",
		Password: "encrypted_password",
		Email:    "test@example.com",
		Roles: []Role{
			{ID: 1, Name: "admin", Permissions: []string{"read", "write"}},
			{ID: 2, Name: "user", Permissions: []string{"read"}},
		},
		Tokens: []Token{
			{ID: 1, Subject: "sub1"},
			{ID: 2, Subject: "sub2"},
		},
		Metadata: map[string]string{
			"dept": "engineering",
		},
	}

	copied := Copy(original)

	g.Expect(copied).To(gomega.Equal(original))
	g.Expect(copied).NotTo(gomega.BeIdenticalTo(original))
	g.Expect(copied.Roles).NotTo(gomega.BeIdenticalTo(original.Roles))
	g.Expect(copied.Tokens).NotTo(gomega.BeIdenticalTo(original.Tokens))
	g.Expect(copied.Metadata).NotTo(gomega.BeIdenticalTo(original.Metadata))

	// Verify complete independence
	copied.Password = "MODIFIED"
	copied.Roles[0].Name = "MODIFIED"
	copied.Roles[0].Permissions[0] = "MODIFIED"
	copied.Tokens[0].Subject = "MODIFIED"
	copied.Metadata["dept"] = "MODIFIED"

	g.Expect(original.Password).To(gomega.Equal("encrypted_password"))
	g.Expect(original.Roles[0].Name).To(gomega.Equal("admin"))
	g.Expect(original.Roles[0].Permissions[0]).To(gomega.Equal("read"))
	g.Expect(original.Tokens[0].Subject).To(gomega.Equal("sub1"))
	g.Expect(original.Metadata["dept"]).To(gomega.Equal("engineering"))
}
