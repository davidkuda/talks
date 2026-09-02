package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var (
	clientID     = os.Getenv("OIDC_CLIENT_ID")
	clientSecret = os.Getenv("OIDC_CLIENT_SECRET")
	issuer       = os.Getenv("OIDC_ISSUER")
	redirectURL  = os.Getenv("OIDC_REDIRECT_URL")
)

type WebApp struct {
	sessions      map[string]string
	pkceVerifiers map[string]string
	users         map[string]User
	OIDC          openIDConnect
}

type User struct {
	SUB          string
	EmailAddress string
}

type openIDConnect struct {
	provider    *oidc.Provider
	verifier    *oidc.IDTokenVerifier
	config      oauth2.Config
	accessToken string
}

func main() {
	webapp := WebApp{}

	webapp.pkceVerifiers = make(map[string]string)
	webapp.sessions = make(map[string]string)
	webapp.users = make(map[string]User)

	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		err = fmt.Errorf("could not instantiate provider: %s", err)
		log.Fatal(err)
	}
	webapp.OIDC.provider = provider

	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	webapp.OIDC.verifier = provider.Verifier(oidcConfig)

	webapp.OIDC.config = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	mux := webapp.routes()
	log.Println("listening on http://localhost:8404/")
	err = http.ListenAndServe("localhost:8404", mux)
	log.Fatal(err)
}

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
