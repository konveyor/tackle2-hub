package auth

import (
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

type LDAP struct {
	URL      string
	OU       string
	BaseDN   string
	Userid   string
	Password string
}

func (r *LDAP) Authenticate() (groupList []string, err error) {
	userDN := fmt.Sprintf(
		"uid=%s,ou=%s,%s",
		r.Userid,
		r.OU,
		r.BaseDN)
	conn, err := ldap.DialURL(r.URL)
	if err != nil {
		return
	}
	defer func() {
		_ = conn.Close()
	}()
	err = conn.Bind(userDN, r.Password)
	if err != nil {
		return
	}
	groupList, err = r.memberOf(conn, userDN)
	if err == nil {
		return
	}
	groupList, err = r.memberSearch(conn, r.BaseDN, userDN)
	if err != nil {
		return
	}
	return
}

func (r *LDAP) memberOf(conn *ldap.Conn, userDN string) (groupList []string, err error) {
	request := ldap.NewSearchRequest(
		userDN,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(objectClass=*)",
		[]string{
			"cn",
			"uid",
			"memberOf"},
		nil,
	)
	result, err := conn.Search(request)
	if err != nil {
		return
	}
	if len(result.Entries) < 1 {
		return
	}
	var pdn *ldap.DN
	memberList := result.Entries[0].GetAttributeValues("memberOf")
	for _, dn := range memberList {
		pdn, err = ldap.ParseDN(dn)
		if err != nil {
			return
		}
		for _, rdn := range pdn.RDNs {
			for _, attr := range rdn.Attributes {
				if strings.ToLower(attr.Type) == "cn" {
					cn := attr.Value
					groupList = append(groupList, cn)
					break
				}
			}
		}
	}
	return
}

func (r *LDAP) memberSearch(conn *ldap.Conn, baseDN, userDN string) (groupList []string, err error) {
	filter := fmt.Sprintf(
		"(&(objectClass=groupOfNames)(member=%s))",
		ldap.EscapeFilter(userDN))
	request := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		[]string{"cn"},
		nil,
	)
	result, err := conn.Search(request)
	if err != nil {
		return
	}
	for _, entry := range result.Entries {
		if cn := entry.GetAttributeValue("cn"); cn != "" {
			groupList = append(groupList, cn)
			break
		}
		if cn := entry.GetAttributeValue("CN"); cn != "" {
			groupList = append(groupList, cn)
			break
		}
	}
	return
}
