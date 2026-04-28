package auth_test

import (
	"context"
	"fmt"

	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/binding/auth"
)

// ExampleOIDC demonstrates complete device authorization flow with Hub.
func ExampleOIDC() {
	hubURL := "https://hub.example.com"

	// Create OIDC authenticator for Hub's OIDC provider
	oidcAuth, err := auth.NewOIDC(hubURL+"/oidc", "konveyor-cli")
	if err != nil {
		panic(err)
	}

	// Perform device login
	err = oidcAuth.DeviceLogin(context.Background())
	if err != nil {
		panic(err)
	}

	// Create Hub API client and set authenticator
	richClient := binding.New(hubURL)
	richClient.Client.Use(oidcAuth)

	// Now use the client - auth headers are automatically set
	// and tokens are automatically refreshed on 401
	apps, err := richClient.Application.List()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d applications\n", len(apps))
}

// ExampleAPIKey demonstrates using a static API key for authentication.
func ExampleAPIKey() {
	hubURL := "https://hub.example.com"

	// Create API key authenticator
	apiKeyAuth := auth.NewAPIKey("my-api-key-token")

	// Create Hub API client and set authenticator
	richClient := binding.New(hubURL)
	richClient.Client.Use(apiKeyAuth)

	// Use the client normally
	apps, err := richClient.Application.List()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d applications\n", len(apps))
}
