package binding

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestConfigMap(t *testing.T) {
	g := NewGomegaWithT(t)

	// LIST: List all configmaps
	list, err := client.ConfigMap.List()
	g.Expect(err).To(BeNil())
	g.Expect(len(list)).To(BeNumerically(">=", 1)) // At least 1 seeded configmap

	// Verify test-config exists in list
	foundTestConfig := false
	for _, cm := range list {
		if cm.Name == "test-config" {
			foundTestConfig = true
			g.Expect(cm.Data).NotTo(BeNil())
			break
		}
	}
	g.Expect(foundTestConfig).To(BeTrue())

	// GET: Retrieve configmap by name
	configMap, err := client.ConfigMap.Get("test-config")
	g.Expect(err).To(BeNil())
	g.Expect(configMap).NotTo(BeNil())
	g.Expect(configMap.Name).To(Equal("test-config"))
	g.Expect(configMap.Data).NotTo(BeNil())

	// Verify data is a map
	dataMap, ok := configMap.Data.(map[string]interface{})
	g.Expect(ok).To(BeTrue())
	g.Expect(dataMap).To(HaveKey("key1"))
	g.Expect(dataMap).To(HaveKey("key2"))
	g.Expect(dataMap).To(HaveKey("database.url"))
	g.Expect(dataMap).To(HaveKey("database.user"))

	// GET KEY: Retrieve specific key from configmap
	value1, err := client.ConfigMap.GetKey("test-config", "key1")
	g.Expect(err).To(BeNil())
	g.Expect(value1).To(Equal("value1"))

	value2, err := client.ConfigMap.GetKey("test-config", "key2")
	g.Expect(err).To(BeNil())
	g.Expect(value2).To(Equal("value2"))

	dbUrl, err := client.ConfigMap.GetKey("test-config", "database.url")
	g.Expect(err).To(BeNil())
	g.Expect(dbUrl).To(Equal("postgresql://localhost:5432/test"))

	dbUser, err := client.ConfigMap.GetKey("test-config", "database.user")
	g.Expect(err).To(BeNil())
	g.Expect(dbUser).To(Equal("testuser"))

	// GET KEY: Try to retrieve non-existent key
	_, err = client.ConfigMap.GetKey("test-config", "non-existent-key")
	g.Expect(err).NotTo(BeNil())

	// GET: Try to retrieve non-existent configmap
	_, err = client.ConfigMap.Get("non-existent-configmap")
	g.Expect(err).NotTo(BeNil())
}
