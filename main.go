package main

import (
	"encoding/json"
	oidc "github.com/coreos/go-oidc"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"os"
)

var (
	providerURI  = os.Getenv("PROVIDER_URL")
	clientID     = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")
)

func main() {

	if !isEnvExist("PROVIDER_URL") {
		log.Fatal("env URL is not set")
	}
	if !isEnvExist("CLIENT_ID") {
		log.Fatal("env CLIENT_ID is not set")
	}
	if !isEnvExist("CLIENT_SECRET") {
		log.Fatal("env CLIENT_SECRET is not set")
	}

	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, providerURI)
	if err != nil {
		log.Fatal(err)
	}

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  "http://localhost:5000/auth/oidc/callback",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "groups", "offline_access"},
	}

	state := "state should be returned unmodified"
	var verifier = provider.Verifier(&oidc.Config{ClientID: clientID})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Redirect to oidc debugger Authorization endpoint")
		http.Redirect(w, r, config.AuthCodeURL(state), http.StatusFound)
	})

	http.HandleFunc("/auth/oidc/callback", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Processing Authorization response")
		if r.URL.Query().Get("state") != state {
			http.Error(w, "state did not match", http.StatusBadRequest)
			return
		}

		oauth2Token, err := config.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			log.Printf("no rawIDTOKEN returned")
		}

		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(oauth2Token))
		if err != nil {
			http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
			return
		}

		resp := struct {
			OAuth2Token *oauth2.Token
			UserInfo    *oidc.UserInfo
			RawIDToken  string
			IDToken     *oidc.IDToken
		}{oauth2Token, userInfo, rawIDToken, idToken}
		data, err := json.MarshalIndent(resp, "", "    ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(data)
	})

	log.Printf("listening on http://%s/", "localhost:5000")
	log.Fatal(http.ListenAndServe("localhost:5000", nil))
}

func isEnvExist(key string) bool {
	if _, ok := os.LookupEnv(key); ok {
		return true
	}
	return false
}
