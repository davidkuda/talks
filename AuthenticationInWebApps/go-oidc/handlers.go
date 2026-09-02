package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type TemplateData struct {
	IsLoggedIn   bool
	EmailAddress string
	SUB          string
}

func (webapp *WebApp) showHomepage(w http.ResponseWriter, r *http.Request) {
	templateData := &TemplateData{
		IsLoggedIn: false,
	}

	if sessionCookie, err := r.Cookie("session"); err == nil {
		sub, sessionExists := webapp.sessions[sessionCookie.Value]
		user, userExists := webapp.users[sub]

		if sessionExists && userExists {
			templateData.IsLoggedIn = true
			templateData.EmailAddress = user.EmailAddress
			templateData.SUB = user.SUB
		}
	}

	t, err := template.New("base").ParseFiles("public/index.html")
	if err != nil {
		fmt.Printf("Error parsing template files: %s\n", err.Error())
		return
	}

	buf := bytes.Buffer{}
	err = t.ExecuteTemplate(&buf, "base", templateData)
	if err != nil {
		err = fmt.Errorf("could not execute template: %s", err)
		log.Println(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	buf.WriteTo(w)
}

func (webapp *WebApp) logout(w http.ResponseWriter, r *http.Request) {
	if sessionCookie, err := r.Cookie("session"); err == nil {
		delete(webapp.sessions, sessionCookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   r.TLS != nil,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (webapp *WebApp) startOIDCFlow(w http.ResponseWriter, r *http.Request) {
	state, err := randString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "state",
		Value:    state,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	})

	nonce, err := randString(16)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "nonce",
		Value:    nonce,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
	})

	pkce := oauth2.GenerateVerifier()
	webapp.pkceVerifiers[state] = pkce

	// TODO: Create redirect URL by hand
	redirectURL := webapp.OIDC.config.AuthCodeURL(
		state,
		oidc.Nonce(nonce),
		oauth2.S256ChallengeOption(pkce),
	)

	prettyPrintURLAsTable(redirectURL)

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (webapp *WebApp) handleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	state, err := r.Cookie("state")
	if err != nil {
		http.Error(w, "state not found", http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("state") != state.Value {
		http.Error(w, "state did not match", http.StatusBadRequest)
		return
	}

	pkceVerifier, ok := webapp.pkceVerifiers[state.Value]
	if !ok {
		http.Error(w, "could not retrieve pkce verifier", http.StatusBadRequest)
		return
	}

	ctx := context.TODO()
	oauth2Token, err := webapp.OIDC.config.Exchange(
		ctx,
		r.URL.Query().Get("code"),
		oauth2.VerifierOption(pkceVerifier),
	)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Authorization codes and their PKCE verifiers are single-use. Removing the
	// verifier also makes an accidental callback refresh fail before contacting
	// the token endpoint a second time.
	delete(webapp.pkceVerifiers, state.Value)
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return
	}
	idToken, err := webapp.OIDC.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	nonce, err := r.Cookie("nonce")
	if err != nil {
		http.Error(w, "nonce not found", http.StatusBadRequest)
		return
	}
	if idToken.Nonce != nonce.Value {
		http.Error(w, "nonce did not match", http.StatusBadRequest)
		return
	}

	var claims struct {
		SUB          string `json:"sub"`
		EmailAddress string `json:"email"`
	}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "Failed to read ID Token claims: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sessionID, err := randString(32)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	user := User{SUB: claims.SUB, EmailAddress: claims.EmailAddress}
	webapp.users[user.SUB] = user
	webapp.sessions[sessionID] = user.SUB

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
