package settings

import (
	"context"

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
	Idp  OpenidProvider
	Ldap LdapProvider
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

	logr.New("auth", 0).Info("Loaded (IdP) resources:\n" + r.String())
	return
}

// String returns a string representation.
func (r Federated) String() (s string) {
	b, _ := yaml.Marshal(r)
	s = string(b)
	return
}

// getIdp get a federated openid provider.
func (r *Federated) getIdp() (err error) {
	list := crd.OpenidProviderList{}
	opt := client2.InNamespace(r.namespace)
	err = r.client.List(context.Background(), &list, opt)
	if err != nil {
		return
	}
	for _, m := range list.Items {
		m2 := OpenidProvider{}
		m2.with(&m)
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
		m2.with(&m)
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

// OpenidProvider defines a federated IdP.
type OpenidProvider struct {
	Enabled      bool
	Name         string
	Issuer       string
	ClientId     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
}

// with populates self with the crd.
func (r *OpenidProvider) with(idp *crd.OpenidProvider) {
	r.Enabled = true
	r.Name = idp.Name
	r.Issuer = idp.Spec.Issuer
	r.ClientId = idp.Spec.ClientId
	r.RedirectURI = idp.Spec.RedirectURI
	r.Scopes = idp.Spec.Scopes
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
}

// with populates self using the crd.
func (r *LdapProvider) with(ds *crd.LdapProvider) {
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
}

// MappingRule defines how LDAP groups are mapped to roles.
type MappingRule struct {
	Any   []string
	And   []string
	Roles []string
}
