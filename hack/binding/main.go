package main

import (
	"context"
	"fmt"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/binding/auth"
)

func main() {
	hubURL := "http://f35a.redhat.com:7070"

	// Create OIDC authenticator
	bearer, err := auth.NewBearer(hubURL+"/oidc", "cli")
	if err != nil {
		panic(err)
	}

	// Perform device login
	err = bearer.DeviceLogin(context.Background())
	if err != nil {
		panic(err)
	}

	// Create Hub API client and set authenticator
	client := binding.New(hubURL)
	client.Client.Use(bearer)

	// Test the client
	apps, err := client.Application.List()
	if err != nil {
		panic(err)
	}

	// Get an apikey.
	pat := &api.PAT{Lifespan: 24}
	err = client.Token.Create(pat)
	if err != nil {
		panic(err)
	}

	bearer.Use(pat.Token)

	// Test the client using the key.
	apps, err = client.Application.List()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Successfully authenticated! Found %d applications\n", len(apps))
}
