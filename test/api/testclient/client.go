package testclient

import (
	"os"

	"github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
)

type RestClient struct {
	client *addon.Client
}

//var restClient, _ = NewHubClient()

// Add with and without login/token creation
func NewHubClient() (client *addon.Client, err error) {
	client = addon.NewClient(os.Getenv("HUB_ENDPOINT"), "")
	token, err := login(client, "admin", "foobar")
	if err != nil {
		return
	}
	client.SetToken(token)
	return
}

type LoginObject struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

// Login performs a login request to the hub and returns a token
func login(client *addon.Client, username, password string) (string, error) {
	login := LoginObject{User: username, Password: password}
	err := client.Post(api.AuthLoginRoot, &login)
	if err != nil {
		return "", err
	}
	return login.Token, nil
}
