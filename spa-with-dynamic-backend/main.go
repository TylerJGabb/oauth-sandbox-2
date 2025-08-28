package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func main() {

	clientId := os.Getenv("OAUTH_CLIENT_ID")
	clientSecret := os.Getenv("OAUTH_CLIENT_SECRET")
	issuer := "https://udemy-tenant-tg.us.auth0.com/"
	redirectUrl := "http://localhost:8080/oauth-callback"

	if clientId == "" || clientSecret == "" {
		log.Fatal("OAUTH_CLIENT_ID and OAUTH_CLIENT_SECRET must be set")
	}

	oidcProvider, err := oidc.NewProvider(context.Background(), issuer)
	if err != nil {
		panic(err)
	}

	oAuthConf := oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
		Endpoint:     oidcProvider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID},
	}

	oidcConfig := oidc.Config{
		ClientID: clientId,
	}

	store, err := redis.NewStore(
		10, // pool size
		"tcp",
		"localhost:6379",   // redis address
		"",                 // redis user
		"",                 // redis password
		[]byte("auth-key"), // is this necessary?
	)

	if err != nil {
		panic(err)
	}

	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60,
		HttpOnly: true,                 // not readable by JS
		Secure:   false,                // if this is deployed behind TLS, set this to true
		SameSite: http.SameSiteLaxMode, // adjust as needed
		// (?) what about wildcards?
		// Domain:  "your.domain", // set if you need a specific domain
	})

	r := gin.Default()
	r.Use(sessions.Sessions("session-storage", store))

	r.GET("/oauth-exchange", func(ctx *gin.Context) {
		code := ctx.Query("code")
		verifier := ctx.Query("verifier")
		redirectUri := ctx.Query("redirectUri")
		if code == "" || verifier == "" || redirectUri == "" {
			log.Default().Println("Missing parameters")
			url := url.URL{
				Path: "/whoops",
				RawQuery: url.Values{
					"error": {"Missing parameters"},
				}.Encode(),
			}
			ctx.Redirect(http.StatusTemporaryRedirect, url.String())
			return
		}
		token, err := oAuthConf.Exchange(
			ctx.Request.Context(),
			code,
			oauth2.VerifierOption(verifier),
			oauth2.SetAuthURLParam("redirect_uri", redirectUri),
		)
		if err != nil {
			log.Default().Printf("Failed to exchange token: %v", err)
			url := url.URL{
				Path: "/whoops",
				RawQuery: url.Values{
					"error": {fmt.Sprintf("Failed to exchange token: %v", err)},
				}.Encode(),
			}
			ctx.Redirect(http.StatusTemporaryRedirect, url.String())
			return
		}

		session := sessions.Default(ctx)
		session.Set("access_token", token.AccessToken)
		session.Set("id_token", token.Extra("id_token"))
		session.Set("refresh_token", token.Extra("refresh_token"))
		session.Save()
		ctx.Redirect(http.StatusTemporaryRedirect, "/")
	})

	r.GET("/whoami", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		idToken, ok := session.Get("id_token").(string)
		if !ok {
			log.Default().Println("No ID token in session")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		verifier := oidcProvider.Verifier(&oidcConfig)
		idTok, err := verifier.Verify(ctx, idToken)
		if err != nil {
			log.Default().Printf("Failed to verify ID token: %v", err)
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		idTokClaims := struct {
			Name string `json:"name"`
		}{}
		if err := idTok.Claims(&idTokClaims); err != nil {
			log.Default().Printf("Failed to extract claims: %v", err)
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		ctx.JSON(http.StatusOK, idTokClaims.Name)
	})

	r.GET("/tokens", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		accessToken := session.Get("access_token")
		if accessToken == nil {
			log.Default().Println("No access token in session")
			url := url.URL{
				Path: "/whoops",
				RawQuery: url.Values{
					"error": {"No access token in session"},
				}.Encode(),
			}
			ctx.Redirect(http.StatusTemporaryRedirect, url.String())
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"access_token":  accessToken,
			"id_token":      session.Get("id_token"),
			"refresh_token": session.Get("refresh_token"),
		})
	})

	r.GET("/logout", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		session.Clear()
		session.Options(sessions.Options{
			Path:     "/",
			MaxAge:   -1, // delete cookie
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})
		session.Save()
		url := url.URL{
			Scheme: "https",
			Host:   "udemy-tenant-tg.us.auth0.com",
			Path:   "/v2/logout",
			RawQuery: url.Values{
				"returnTo": {"http://localhost:8080"},
			}.Encode(),
		}
		ctx.Redirect(http.StatusTemporaryRedirect, url.String())
	})

	r.GET("/whoops", func(ctx *gin.Context) {
		err := ctx.Query("error")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
	})

	r.GET("/", func(ctx *gin.Context) {
		ctx.File("spa-with-dynamic-backend/index.html")
	})
	r.Run(":8080")

}
