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

	r.GET("/login", func(ctx *gin.Context) {
		name := ctx.Query("name")
		verifier := oauth2.GenerateVerifier()

		session := sessions.Default(ctx)
		session.Set("user", name)
		session.Set("verifier", verifier)
		session.Save()

		stateNotUsed := "state_not_used"
		redirectUrl := oAuthConf.AuthCodeURL(stateNotUsed,
			oauth2.S256ChallengeOption(verifier),
			oauth2.SetAuthURLParam("audience", "https://contacts.example.com"),
		)
		ctx.Redirect(http.StatusTemporaryRedirect, redirectUrl)
	})

	r.GET("/oauth-callback", func(ctx *gin.Context) {
		code := ctx.Query("code")
		if code == "" {
			log.Default().Println("The authorization server did not send a code")
			url := url.URL{
				Path: "/whoops",
				RawQuery: url.Values{
					"error": {"The authorization server did not send a code"},
				}.Encode(),
			}
			ctx.Redirect(http.StatusTemporaryRedirect, url.String())
			return
		}

		session := sessions.Default(ctx)
		if session == nil {
			ctx.Redirect(http.StatusTemporaryRedirect, "/login")
			return
		}

		verifier, ok := session.Get("verifier").(string)
		if !ok {
			log.Default().Printf("Failed to get verifier from session")
			url := url.URL{
				Path: "/whoops",
				RawQuery: url.Values{
					"error": {"Failed to get verifier from session"},
				}.Encode(),
			}
			ctx.Redirect(http.StatusTemporaryRedirect, url.String())
			return
		}

		token, err := oAuthConf.Exchange(
			ctx.Request.Context(),
			code,
			oauth2.VerifierOption(verifier),
		)

		if err != nil {
			log.Default().Printf("Failed to exchange token: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
			url := url.URL{
				Path: "/whoops",
				RawQuery: url.Values{
					"error": {fmt.Sprintf("Failed to exchange token: %v", err)},
				}.Encode(),
			}
			ctx.Redirect(http.StatusTemporaryRedirect, url.String())
			return
		}

		session.Set("access_token", token.AccessToken)
		session.Save()
		ctx.Redirect(http.StatusTemporaryRedirect, "/profile")
	})

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
		session.Save()
		ctx.Redirect(http.StatusTemporaryRedirect, "/profile")
	})

	r.GET("/profile", func(ctx *gin.Context) {
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
			"access_token": accessToken,
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
		ctx.JSON(http.StatusOK, gin.H{"message": "Logged out"})
	})

	r.GET("/whoops", func(ctx *gin.Context) {
		err := ctx.Query("error")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
	})

	r.GET("/", func(ctx *gin.Context) {
		ctx.File("client/index.html")
	})
	r.Run(":8080")

}
