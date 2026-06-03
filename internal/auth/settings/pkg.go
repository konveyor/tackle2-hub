package settings

import (
	"context"
	"crypto/tls"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/internal/k8s"
	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
	"gopkg.in/yaml.v3"
	core "k8s.io/api/core/v1"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"
)

// Federated defines federation settings.
type Federated struct {
	Idp     IdentityProvider
	Ldap    LdapProvider
	Clients []IdpClient
	//
	client    client2.Client
	namespace string
}

// Load settings.
func (r *Federated) Load(namespace string) (err error) {
	r.namespace = namespace

	r.client, err = k8s.NewClient()
	if err != nil {
		return
	}

	err = r.getIdp()
	if err != nil {
		return
	}
	err = r.getLdap()
	if err != nil {
		return
	}
	err = r.getClients()
	if err != nil {
		return
	}

	logr.New("auth", 0).Info("Loaded (IdP) resources:\n" + r.String())
	return
}

// String returns a string representation.
func (r Federated) String() (s string) {
	r.Idp.TLS = nil
	r.Ldap.TLS = nil
	b, _ := yaml.Marshal(r)
	s = string(b)
	return
}

// getIdp get a federated identity provider.
func (r *Federated) getIdp() (err error) {
	list := crd.IdentityProviderList{}
	opt := client2.InNamespace(r.namespace)
	err = r.client.List(context.Background(), &list, opt)
	if err != nil {
		return
	}
	for _, m := range list.Items {
		m2 := IdentityProvider{}
		err = m2.with(&m)
		if err != nil {
			return
		}
		ref := m.Spec.ClientSecret
		if ref != nil {
			m2.ClientSecret, err = r.secret(ref, "clientSecret")
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}
		r.Idp = m2
		break
	}
	return
}

// getLdap get a federated LDAP provider.
func (r *Federated) getLdap() (err error) {
	list := crd.LdapProviderList{}
	opt := client2.InNamespace(r.namespace)
	err = r.client.List(context.Background(), &list, opt)
	if err != nil {
		return
	}
	for _, m := range list.Items {
		m2 := LdapProvider{}
		err = m2.with(&m)
		if err != nil {
			return
		}
		ref := m.Spec.Password
		if ref != nil {
			m2.Password, err = r.secret(ref, "password")
			if err != nil {
				return
			}
		}
		r.Ldap = m2
		break
	}
	return
}

// getClients gets federated OIDC clients.
func (r *Federated) getClients() (err error) {
	list := crd.IdpClientList{}
	opt := client2.InNamespace(r.namespace)
	err = r.client.List(context.Background(), &list, opt)
	if err != nil {
		return
	}

	for _, m := range list.Items {
		client := IdpClient{
			ID:              m.Spec.ID,
			ClientId:        m.Spec.ClientId,
			ApplicationType: m.Spec.ApplicationType,
			Grants:          m.Spec.Grants,
			RedirectURIs:    m.Spec.RedirectURIs,
			Scopes:          m.Spec.Scopes,
		}
		ref := m.Spec.ClientSecret
		if ref != nil {
			client.Secret, err = r.secret(ref, "clientSecret")
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}

		r.Clients = append(r.Clients, client)
	}
	return
}

// secret returns the secret by key.
func (r *Federated) secret(ref *core.ObjectReference, key string) (s string, err error) {
	secret := &core.Secret{}
	err = r.client.Get(
		context.Background(),
		client2.ObjectKey{
			Namespace: ref.Namespace,
			Name:      ref.Name,
		},
		secret)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	b := secret.Data[key]
	s = string(b)
	return
}

// IdentityProvider defines a federated IdP.
type IdentityProvider struct {
	Enabled      bool
	Primary      bool
	Name         string
	Issuer       string
	ClientId     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
	TLS          *tls.Config
}

// with populates self with the crd.
func (r *IdentityProvider) with(idp *crd.IdentityProvider) (err error) {
	r.Enabled = true
	r.Primary = idp.Spec.Primary
	r.Name = idp.Name
	r.Issuer = idp.Spec.Issuer
	r.ClientId = idp.Spec.ClientId
	r.RedirectURI = idp.Spec.RedirectURI
	r.Scopes = idp.Spec.Scopes
	r.TLS, err = idp.Spec.TLS.AsConfig()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// LdapProvider defines a federated LDAP directory.
type LdapProvider struct {
	Enabled      bool
	Name         string
	Kind         string
	URL          string
	BaseDN       string
	BindDN       string
	Password     string
	UserFilter   string
	GroupFilter  string
	HasMemberOf  bool
	RoleMappings []MappingRule
	TLS          *tls.Config
}

// with populates self using the crd.
func (r *LdapProvider) with(ds *crd.LdapProvider) (err error) {
	r.Enabled = true
	r.Name = ds.Name
	r.Kind = ds.Spec.Kind
	r.URL = ds.Spec.URL
	r.BaseDN = ds.Spec.BaseDN
	r.BindDN = ds.Spec.BindDN
	r.UserFilter = ds.Spec.UserFilter
	r.GroupFilter = ds.Spec.GroupFilter
	r.HasMemberOf = ds.Spec.HasMemberOf
	for _, m := range ds.Spec.RoleMappings {
		r.RoleMappings = append(r.RoleMappings, MappingRule{
			Any:   m.Any,
			And:   m.And,
			Roles: m.Roles,
		})
	}
	r.TLS, err = ds.Spec.TLS.AsConfig()
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

// MappingRule defines how LDAP groups are mapped to roles.
type MappingRule struct {
	Any   []string
	And   []string
	Roles []string
}

// IdpClient represents a loaded OIDC client configuration.
type IdpClient struct {
	ID              uint
	ClientId        string
	Secret          string
	ApplicationType string
	Grants          []string
	RedirectURIs    []string
	Scopes          []string
}
