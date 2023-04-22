package utils

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/crane"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

type RegistryCreds struct {
	Username string
	Password string
	Registry string
}

func parseRegistryCreds(input string) (*RegistryCreds, error) {
	// Split the input into username:password, and registry parts
	parts := strings.Split(input, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid registry credentials input format")
	}
	userpass := parts[0]
	registry := parts[1]

	// Parse the username and password
	userparts := strings.Split(userpass, ":")
	if len(userparts) != 2 {
		return nil, fmt.Errorf("invalid username:password format")
	}
	username := userparts[0]
	password := userparts[1]

	// Parse the registry URL and make sure it's valid
	registryURL, err := url.Parse("https://" + registry)
	if err != nil {
		return nil, fmt.Errorf("invalid registry URL: %v", err)
	}
	if registryURL.Scheme != "https" {
		return nil, fmt.Errorf("registry must use HTTPS")
	}

	return &RegistryCreds{
		Username: username,
		Password: password,
		Registry: registryURL.Host,
	}, nil
}

func getAuthOpt(creds RegistryCreds) (crane.Option, error) {
	basicAuth := authn.Basic{
		Username: creds.Username,
		Password: creds.Password,
	}
	authenticator, err := basicAuth.Authorization()
	if err != nil {
		return nil, err
	}
	authConfig := authn.FromConfig(*authenticator)
	return crane.WithAuth(authConfig), nil
}

func PushImage(image v1.Image, registryCredits string, dest string) error {
	opts := []crane.Option{}
	if registryCredits != "default" {
		creds, err := parseRegistryCreds(registryCredits)
		if err != nil {
			return err
		}
		dest = creds.Registry + "/" + dest
		authOpt, err := getAuthOpt(*creds)
		if err != nil {
			return err
		}
		opts = append(opts, authOpt)
	}
	return crane.Push(image, dest, opts...)
}
