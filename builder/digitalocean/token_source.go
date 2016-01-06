package digitalocean

import (
	"golang.org/x/oauth2"
)

type apiTokenSource struct {
	AccessToken string
}

func (t *apiTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: t.AccessToken,
	}, nil
}
