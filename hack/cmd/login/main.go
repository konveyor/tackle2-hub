package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/binding/auth"
)

// Login CLI: Login
// Examples:
//   $ login -r http://route.konveyor.io
//   $ login -u http://localhost:8080
//   $ login -u http://localhost:8080 -i http://localhost:8080/oidc

const (
	ClientId = "kantra"
)

// AuthMethod provides an example of how to extend
// The OIDC auth method to automatically upgrade the token
// to an APIKey (token) with the specified lifespan.
// Header() is called for every request and provides an
// opportunity to read a stored token.  Login() is called when
// the binding client determines new token is needed. After login
// succeeds, an api-key is requested using a separate binding. The token
// is set in the embedded OIDC method and stored.
type AuthMethod struct {
	*auth.OIDC
	hubURL    string
	lifespan  int
	transport *http.Transport
}

// Header returns the authorization header.
// Reads the stored PAT/api-key as needed.
func (m *AuthMethod) Header() (h string) {
	if m.Token() == "" {
		m.Use(m.getToken())
	}
	h = m.OIDC.Header()
	return
}

// Login called as needed by the binding to get a new token.
func (m *AuthMethod) Login() (err error) {
	err = m.OIDC.Login()
	if err != nil {
		return
	}

	richClient := binding.New(m.hubURL)
	richClient.Client.SetTransport(m.transport)
	richClient.Client.Use(auth.NewBearer(m.OIDC.Token()))
	key := &api.APIKey{Lifespan: m.lifespan}
	err = richClient.Token.Create(key)
	if err != nil {
		return
	}

	m.Use(key.Token)
	m.storeToken(key.Token)
	return
}

// getToken returns the stored api-key.
func (m *AuthMethod) getToken() (key string) {
	return
}

// storeToken stores the api-key.
func (m *AuthMethod) storeToken(key string) {
}

// SetTransport sets the (optional) HTTP transport.
func (m *AuthMethod) SetTransport(tr *http.Transport) {
	m.transport = tr
	m.OIDC.SetTransport(tr)
}

func main() {
	defer func() {
		p := recover()
		if p == nil {
			return
		}
		if err, cast := p.(error); cast {
			fmt.Println(err.Error())
		} else {
			fmt.Printf("\n%v\n", p)
		}
	}()

	route := flag.String(
		"r",
		"",
		"Openshift/kubernetes route (ingress URL.")
	hubURL := flag.String(
		"u",
		"http://localhost:8080",
		"Hub base URL.")
	issuerURL := flag.String(
		"i",
		"",
		"OIDC issuer URL.")
	clientId := flag.String(
		"c",
		ClientId,
		"OIDC client ID.")
	patLifespan := flag.Int(
		"h",
		720,
		"PAT lifespan (hours) Default: 30 days.")
	token := flag.String(
		"b",
		"",
		"Use bearer token (apikey).")
	login := flag.String(
		"login",
		"",
		"Login (non-federated OIDC user).")
	password := flag.String(
		"password",
		"",
		"User password.")
	flag.Parse()

	if *issuerURL == "" {
		var path string
		if *route != "" {
			path, _ = url.JoinPath(*route, "oidc")
		} else {
			path, _ = url.JoinPath(*hubURL, "oidc")
		}
		issuerURL = &path
	}
	if *route != "" {
		hubURL = route
	}

	fmt.Printf("\nUsing:\n")
	fmt.Printf("  route:     %s\n", *route)
	fmt.Printf("  hubURL:    %s\n", *hubURL)
	fmt.Printf("  issuerURL: %s\n", *issuerURL)
	fmt.Printf("  clientId:  %s\n", *clientId)
	fmt.Printf("  token:     %s\n", *token)
	fmt.Printf("  login:     %s\n", *login)
	fmt.Printf("  password:  %s\n", *password)
	fmt.Printf("\n")

	// API client.
	richClient := binding.New(*hubURL)

	// OIDC authentication.
	if *login == "" && *token == "" {
		tr := richClient.Client.Transport()
		tr.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}

		authMethod := &AuthMethod{
			OIDC:     auth.NewOIDC(*issuerURL, *clientId),
			hubURL:   *hubURL,
			lifespan: *patLifespan,
		}
		authMethod.SetTransport(tr)
		err := authMethod.Login()
		if err != nil {
			panic(err)
		}

		fmt.Printf("\nAuth succeeded. token: %s\n", authMethod.Token())

		richClient.UseAuth(authMethod)

		testClient(richClient)
		return
	}

	// User supplied token.
	if *token != "" {
		bearer := &auth.OIDC{}
		bearer.Use(*token)
		richClient.Client.Use(bearer)
		testClient(richClient)
	}

	// Basic auth.
	if *login != "" {
		basic := auth.NewBasic(*login, *password)
		richClient.Client.Use(basic)
		testClient(richClient)
	}

	println("Done")
}

func testClient(richClient *binding.RichClient) {
	_, err := richClient.User.List()
	if err != nil {
		panic(err)
	}
}
