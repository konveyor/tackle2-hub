package client

import (
	"os"

	"github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
)

type Client struct {
	client *addon.Client
}

// Add with and without login/token creation
func NewHubClient() (client *addon.Client, err error) {
	client = addon.NewClient(os.Getenv("HUB_BASE_URL"), "")
	token, err := login(client, "admin", "foobar")
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
