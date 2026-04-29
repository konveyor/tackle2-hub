package main

import (
	"context"
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
	ClientId = "cli"
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
		24,
		"PAT lifespan (hours).")
	token := flag.String(
		"b",
		"",
		"Use bearer token (apikey).")
	userid := flag.String(
		"userid",
		"",
		"User ID (non-federated OIDC user).")
	password := flag.String(
		"password",
		"",
		"User password.")
	flag.Parse()

	if *issuerURL == "" {
		var path string
		if *route != "" {
			path, _ = url.JoinPath(*route, "auth")
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
	fmt.Printf("  token:     %s\n", *token)
	fmt.Printf("  userid:    %s\n", *userid)
	fmt.Printf("  password:  %s\n", *password)
	fmt.Printf("\n")

	// API client.
	richClient := binding.New(*hubURL)

	// OIDC authentication.
	if *userid == "" && *token == "" {

		bearer, err := auth.NewBearer(*issuerURL, *clientId)
		if err != nil {
			panic(err)
		}
		err = bearer.DeviceLogin(context.Background())
		if err != nil {
			panic(err)
		}

		fmt.Printf("\nAuth succeeded. token: %s\n", bearer.Token())

		richClient.Client.Use(bearer)

		pat := &api.PAT{Lifespan: *patLifespan}
		err = richClient.Token.Create(pat)
		if err != nil {
			panic(err)
		}

		bearer.Use(pat.Token)

		fmt.Printf("\nGet PAT succeeded. token: %s\n", bearer.Token())

		testClient(richClient)
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
	if *userid != "" {
		basic := auth.NewBasic(*userid, *password)
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
