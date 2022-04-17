package cmdutils

import (
	"context"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
	"net/http"
	"os"
)

func GetGraphqlClient(ctx context.Context, url string) *graphql.Client {
	accessToken := os.Getenv("ACCESS_TOKEN")
	var httpClient *http.Client
	if accessToken != "" {
		src := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: accessToken},
		)
		httpClient = oauth2.NewClient(ctx, src)
	}
	gqlClient := graphql.NewClient(url, httpClient)
	return gqlClient
}
