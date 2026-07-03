package auth

import (
	"crypto/tls"
	"fmt"
	"sort"
	"strings"
	"time"

	dr "github.com/bmatcuk/doublestar/v4"
	"github.com/go-ldap/ldap/v3"
	liberr "github.com/jortel/go-utils/error"
	"github.com/konveyor/tackle2-hub/internal/auth/settings"
	"github.com/konveyor/tackle2-hub/internal/secret"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LdapHandler provides LDAP authentication.
type LdapHandler struct {
	cache   *Cache
	db      *gorm.DB
	enabled bool
	ds      LDAP
}

// Authenticate a user.
// Authenticated users (identities) are cached.
// Authenticated with the DS when cached identity not found or expired.
// The LDAP is cloned for each concurrent request.
func (h *LdapHandler) Authenticate(login, password string) (subject *Subject, err error) {
	if !h.enabled {
		err = &NotFound{
			Resource: "User",
			Id:       login,
		}
		return
	}
	ds := h.ds // cloned.
	ldapUser, err := ds.Authenticate(login, password)
	if err != nil {
		return
	}
	identity := h.buildIdentity(ldapUser)
	err = h.ensureIdentity(identity)
	if err != nil {
		return
	}
	subject = &Subject{}
	subject.WithIdentity(identity)
	return
}

// buildIdentity builds an IdpIdentity from LDAP user data.
func (h *LdapHandler) buildIdentity(ldapUser *LdapUser) (identity *Identity) {
	var scopes []string
	for _, roleName := range ldapUser.Roles {
		role, err := h.cache.FindRoleByName(roleName)
		if err != nil {
			continue
		}
		scopes = append(scopes, role.Scopes...)
	}
	scopes = uniqueStrings(scopes)
	sort.Strings(scopes)

	identity = &Identity{
		Kind:    IdentityKindLDAP,
		Issuer:  h.ds.URL,
		Subject: ldapUser.Subject,
		Login:   ldapUser.Login,
		Scopes:  scopes,
	}

	return
}

// ensureIdentity ensures the identity is created or updated in the database.
func (h *LdapHandler) ensureIdentity(identity *Identity) (err error) {
	db := h.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "subject"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"kind",
			"issuer",
			"login",
			"name",
			"email",
			"scopes",
			"updateUser",
		}),
	})
	err = db.Create(identity).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	h.cache.IdentitySaved(identity)
	return
}

// LDAP provides LDAP authentication.
type LDAP struct {
	Kind        string
	Name        string
	URL         string
	BaseDN      string
	BindDN      string
	Password    string
	UserFilter  string
	GroupFilter string
	HasMemberOf bool
	TLS         *tls.Config
	//
	mapper RoleMapper
	conn   *ldap.Conn
}

// Authenticate authenticates a user against LDAP and returns group membership.
func (r *LDAP) Authenticate(login, password string) (dsUser *LdapUser, err error) {
	r.Kind = strings.ToUpper(r.Kind)
	if r.TLS != nil {
		r.conn, err = ldap.DialURL(r.URL, ldap.DialWithTLSConfig(r.TLS))
	} else {
		r.conn, err = ldap.DialURL(r.URL)
	}
	if err != nil {
		Log.Info(
			"LDAP connection failed.",
			"reason",
			err.Error())
		return
	}
	defer func() {
		_ = r.conn.Close()
	}()

	// Bind using SA.
	err = r.conn.Bind(r.BindDN, r.Password)
	if err != nil {
		Log.Info(
			"LDAP (SA) bind failed.",
			"dn",
			r.BindDN,
			"reason",
			err.Error())
		return
	}

	// Find the user and authenticate using bind.
	user, err := r.findUser(login)
	if err != nil {
		return
	}
	err = r.conn.Bind(user.DN, password)
	if err != nil {
		err = &NotAuthenticated{
			Reason: "invalid password",
			Token:  login,
		}
	}

	// Find roles
	var groups []string

	defer func() {
		dsUser = &LdapUser{
			DN:      user.DN,
			Login:   login,
			Subject: secret.Hash(login),
			Roles:   r.mapper.roles(groups),
		}
	}()

	// Using memberOf.
	if r.HasMemberOf {
		groups = r.memberOf(user)
		if len(groups) > 0 {
			return
		}
	}

	// Bind using SA.
	err = r.conn.Bind(r.BindDN, r.Password)
	if err != nil {
		Log.Info(
			"LDAP (SA) bind failed.",
			"dn",
			r.BindDN,
			"reason",
			err.Error())
		return
	}
	// Find (search) groups.
	groups, err = r.findGroup(user)
	if err != nil {
		return
	}
	return
}

// findUser finds and returns a user.
func (r *LDAP) findUser(login string) (entry *ldap.Entry, err error) {
	baseDN := fmt.Sprintf("ou=people,%s", r.BaseDN)
	request := r.request(
		baseDN,
		r.userFilter(login),
		"dn",
		"cn",
		"uid",
		"memberOf")
	result, err := r.conn.Search(request)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	if len(result.Entries) == 0 || len(result.Entries) > 1 {
		err = &NotFound{
			Resource: "user",
			Filter:   request.Filter,
			Id:       login,
		}
		return
	}
	entry = result.Entries[0]
	return
}

// findGroup performs a search and returns the groups.
func (r *LDAP) findGroup(user *ldap.Entry) (groups []string, err error) {
	baseDN := fmt.Sprintf("ou=groups,%s", r.BaseDN)
	request := r.request(
		baseDN,
		r.groupFilter(user),
		"dn",
		"cn",
		"mail",
		"memberOf")
	result, err := r.conn.Search(request)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, entry := range result.Entries {
		v := entry.GetAttributeValue("cn")
		if v != "" {
			groups = append(groups, v)
		}
	}
	if len(groups) == 0 {
		Log.Info(
			"LDAP: no groups found using filter.",
			"filter",
			request.Filter)
	}
	return
}

// userFilter returns the user search filter.
// Supported (macros):
// - ${uid} - user uid
// - ${login} - user login (cn alias)
func (r *LDAP) userFilter(login string) (filter string) {
	filter = r.UserFilter
	switch r.Kind {
	case "ACTIVEDIRECTORY",
		"AD":
		if filter == "" {
			filter = "(sAMAccountName=${login})"
		}
	default:
		if filter == "" {
			filter = "(uid=${uid})"
		}
	}

	uid := ldap.EscapeFilter(login)
	filter = strings.Replace(filter, "${login}", uid, -1)
	filter = strings.Replace(filter, "${uid}", uid, -1)

	return
}

// groupFilter returns the group search filter.
// Supported (macros):
// - ${uid} - user uid (login alias)
// - ${cn}  - user cn.
// - ${dn}  - user dn.
func (r *LDAP) groupFilter(user *ldap.Entry) (filter string) {
	filter = r.GroupFilter
	switch r.Kind {
	case "ACTIVEDIRECTORY",
		"AD":
		if filter == "" {
			filter = "(&(objectClass=group)(member=${dn}))"
		}
	default:
		if filter == "" {
			filter = "(&(objectClass=*)(member=${dn}))"
		}
	}

	dn := ldap.EscapeFilter(user.DN)
	uid := ldap.EscapeFilter(user.GetAttributeValue("uid"))
	cn := ldap.EscapeFilter(user.GetAttributeValue("cn"))
	filter = strings.Replace(filter, "${uid}", uid, -1)
	filter = strings.Replace(filter, "${cn}", cn, -1)
	filter = strings.Replace(filter, "${dn}", dn, -1)

	return
}

// request returns an LDAP search request.
func (r *LDAP) request(dn, filter string, attr ...string) (req *ldap.SearchRequest) {
	req = ldap.NewSearchRequest(
		dn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		attr,
		nil,
	)
	return
}

// memberOf returns the value of the memberOf attribute
func (r *LDAP) memberOf(entry *ldap.Entry) (groups []string) {
	values := entry.GetAttributeValues("memberOf")
	for _, dnStr := range values {
		dn, err := ldap.ParseDN(dnStr)
		if err != nil {
			continue
		}
		for _, rdn := range dn.RDNs {
			for _, attr := range rdn.Attributes {
				if strings.EqualFold(attr.Type, "cn") {
					groups = append(groups, attr.Value)
					break
				}
			}
		}
	}
	return
}

// MappingRule defines rules for mapping LDAP groups to roles.
// The And (group) patterns must match ALL.
// The Any (group) patterns must match at least one.
// The pattern is a gob.
type MappingRule struct {
	And   []string
	Any   []string
	Roles []string
}

// Empty returns true when both And and Any patterns are empty.
func (m *MappingRule) Empty() (empty bool) {
	n := 0
	n += len(m.And)
	n += len(m.Any)
	empty = n == 0
	return
}

// RoleMapper provides LDAP group mapping to roles.
type RoleMapper struct {
	RuleSet []MappingRule
}

// Use settings ruleset.
func (r *RoleMapper) Use(ruleSet []settings.MappingRule) {
	for _, rule := range ruleSet {
		r.RuleSet = append(
			r.RuleSet,
			MappingRule{
				And:   rule.And,
				Any:   rule.Any,
				Roles: rule.Roles,
			})
	}
}

// roles maps LDAP groups to hub roles using the configured mapping rules.
func (m *RoleMapper) roles(groups []string) (roles []string) {
	for _, rule := range m.RuleSet {
		if rule.Empty() {
			continue
		}
		andMatched := m.and(rule.And, groups)
		anyMatched := m.any(rule.Any, groups)
		if andMatched && anyMatched {
			roles = append(roles, rule.Roles...)
		}
	}
	roles = uniqueStrings(roles)
	sort.Strings(roles)

	if len(roles) == 0 {
		Log.Info(
			"LDAP: WARNING: No roles matched.",
			"groups",
			groups)
	}
	return
}

// any returns true when any of the patterns are matched.
func (m *RoleMapper) any(patterns, groups []string) (matched bool) {
	if len(patterns) == 0 {
		matched = true
		return
	}
	for _, p := range patterns {
		for _, cn := range groups {
			matched, _ = dr.Match(p, cn)
			if matched {
				return
			}
		}
	}
	return
}

// and returns true when ALL of the patterns have matched.
func (m *RoleMapper) and(patterns, groups []string) (matched bool) {
	n := 0
	if len(patterns) == 0 {
		matched = true
		return
	}
	for _, p := range patterns {
		for _, cn := range groups {
			match, _ := dr.Match(p, cn)
			if match {
				n++
				break
			}
		}
	}
	matched = n == len(patterns)
	return
}

// LdapUser defines an authenticated user.
type LdapUser struct {
	DN         string
	Login      string
	Subject    string
	Roles      []string
	Scopes     []string
	Expiration time.Time
}
