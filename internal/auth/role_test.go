package auth

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/onsi/gomega"
)

func TestLoadYaml(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	roles, err := LoadRoles("./roles.yaml")
	g.Expect(err).To(gomega.BeNil())
	users, err := LoadUsers("./users.yaml")
	g.Expect(err).To(gomega.BeNil())

	validate := validator.New()
	var roleNames []string
	for _, role := range roles {
		err = validate.Struct(role)
		g.Expect(err).To(gomega.BeNil())
		for _, resource := range role.Resources {
			err = validate.Struct(resource)
			g.Expect(err).To(gomega.BeNil())
		}
		roleNames = append(roleNames, role.Name)
	}

	for _, user := range users {
		err = validate.Struct(user)
		g.Expect(err).To(gomega.BeNil())
		for _, role := range user.Roles {
			g.Expect(role).To(gomega.BeElementOf(roleNames))
		}
	}
}
