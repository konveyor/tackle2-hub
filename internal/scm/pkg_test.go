package scm

import (
	"testing"

	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/internal/secret"
	. "github.com/onsi/gomega"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// TestProxyMapDecryptsIdentity tests that the proxy identity
// credentials are decrypted when the proxy map is built.
func TestProxyMapDecryptsIdentity(t *testing.T) {
	g := NewGomegaWithT(t)
	secret.Settings.Passphrase = "test"

	db, err := gorm.Open(
		sqlite.Open("file::memory:?_foreign_keys=yes"),
		&gorm.Config{
			NamingStrategy: &schema.NamingStrategy{
				SingularTable: true,
				NoLowerCase:   true,
			},
		})
	g.Expect(err).To(BeNil())
	err = db.AutoMigrate(&model.Identity{}, &model.Proxy{})
	g.Expect(err).To(BeNil())

	// Stored the same way the identity handler stores it.
	id := &model.Identity{
		Kind:     "proxy",
		Name:     "proxy-id",
		User:     "elmer",
		Password: "rabbit/slayer+1=",
		Key:      "hunting-season",
	}
	err = secret.Encrypt(id)
	g.Expect(err).To(BeNil())
	g.Expect(id.Password).NotTo(Equal("rabbit/slayer+1="))
	err = db.Create(id).Error
	g.Expect(err).To(BeNil())
	proxy := &model.Proxy{
		Enabled:    true,
		Kind:       "http",
		Host:       "proxy.example.com",
		Port:       3128,
		IdentityID: &id.ID,
	}
	err = db.Create(proxy).Error
	g.Expect(err).To(BeNil())

	pm, err := proxyMap(db)
	g.Expect(err).To(BeNil())
	p, found := pm["http"]
	g.Expect(found).To(BeTrue())
	g.Expect(p.Identity).NotTo(BeNil())
	g.Expect(p.Identity.User).To(Equal("elmer"))
	g.Expect(p.Identity.Password).To(Equal("rabbit/slayer+1="))
	g.Expect(p.Identity.Key).To(Equal("hunting-season"))
}
