package main

import (
	"context"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/binding/auth"
)

func main() {
	hubURL := "http://f35a.redhat.com:7070"

	// Create OIDC authenticator
	bearer, err := auth.NewBearer(hubURL, "cli")
	if err != nil {
		panic(err)
	}

	err = bearer.DeviceLogin(context.Background())
	if err != nil {
		panic(err)
	}

	client := binding.New(hubURL)
	client.Client.Use(bearer)

	// Test the client
	_, err = client.User.List()
	if err != nil {
		panic(err)
	}

	// Get a (PAT) personal access token.
	pat := &api.PAT{Lifespan: 24}
	err = client.Token.Create(pat)
	if err != nil {
		panic(err)
	}

	bearer.Use(pat.Token)

	// Test the client using the apikey.
	_, err = client.User.List()
	if err != nil {
		panic(err)
	}

	print("Authentication succeeded.")
}
