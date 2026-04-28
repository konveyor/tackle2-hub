package main

import (
	"context"
	"fmt"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/binding/auth"
)

func main() {
	hubURL := "http://f35a.redhat.com:8080"

	// Create OIDC authenticator
	clientAuth, err := auth.NewOIDC(hubURL+"/oidc", "cli")
	if err != nil {
		panic(err)
	}

	// Perform device login
	err = clientAuth.DeviceLogin(context.Background())
	if err != nil {
		panic(err)
	}

	// Create Hub API client and set authenticator
	client := binding.New(hubURL)
	client.Client.Use(clientAuth)

	// Test the client
	apps, err := client.Application.List()
	if err != nil {
		panic(err)
	}

	pat := &api.PAT{Lifespan: 24}
	err = client.Token.Create(pat)
	if err != nil {
		panic(err)
	}

	clientAuth.Use(pat.Token)

	fmt.Printf("Successfully authenticated! Found %d applications\n", len(apps))
}
