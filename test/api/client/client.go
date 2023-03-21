package client

import (
	"fmt"
	"os"

	"github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
)

var Client *addon.Client

func init() {
	var err error
	baseUrl := os.Getenv("HUB_BASE_URL")
	Client, err = NewHubClient(baseUrl, "admin", "")
	if err != nil {
		panic(fmt.Sprintf("Error: Cannot setup API client for URL '%s': %v.", baseUrl, err.Error()))
	}
}

// Add with and without login/token creation
func NewHubClient(baseUrl, username, password string) (client *addon.Client, err error) {
	client = addon.NewClient(baseUrl, "")
	token, err := login(client, username, password)
	if err != nil {
		return
	}
	client.SetToken(token)
	return
}

// Login performs a login request to the hub and returns a token
func login(client *addon.Client, username, password string) (string, error) {
	login := api.Login{User: username, Password: password}
	err := client.Post(api.AuthLoginRoot, &login)
	if err != nil {
		return "", err
	}
	return login.Token, nil
}
