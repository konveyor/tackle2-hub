package main

import (
	"crypto/tls"
	"flag"
	"fmt"
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
	getToken := flag.Bool(
		"getToken",
		true,
		"Get PAT token.")
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

		bearer := auth.NewBearer(*issuerURL, *clientId)
		bearer.SetTransport(tr)
		err := bearer.Login()
		if err != nil {
			panic(err)
		}

		fmt.Printf("\nAuth succeeded. token: %s\n", bearer.Token())

		richClient.UseAuth(bearer)

		if *getToken {
			pat := &api.PAT{Lifespan: *patLifespan}
			err = richClient.Token.Create(pat)
			if err != nil {
				panic(err)
			}

			bearer.Use(pat.Token)

			fmt.Printf("\nGet PAT succeeded. token: %s\n", bearer.Token())

			testClient(richClient)
		}
		return
	}

	// User supplied token.
	if *token != "" {
		bearer := &auth.Bearer{}
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
