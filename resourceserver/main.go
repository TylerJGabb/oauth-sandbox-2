package main

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/gin-gonic/gin"
	jwt4 "github.com/golang-jwt/jwt/v4"
)

func main() {

	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// todo: somehow dynamically detect this from the token iss
	// cache all jwks created for each issuer, use cached issuer, if available.
	jwks, err := keyfunc.Get("https://udemy-tenant-tg.us.auth0.com/.well-known/jwks.json", keyfunc.Options{
		RefreshUnknownKID: true,
		RefreshErrorHandler: func(err error) {
			log.Printf("There was an error with the jwt.Keyfunc: %s", err.Error())
		},
	})

	if err != nil {
		log.Fatalf("Failed to build keyfunc: %s", err.Error())
	}

	r.GET("/resource", func(c *gin.Context) {
		bearerToken := c.Request.Header.Get("Authorization")
		if bearerToken == "" {
			c.JSON(401, gin.H{"error": "Missing Authorization Header"})
			return
		}

		accessToken := strings.TrimPrefix(bearerToken, "Bearer ")
		jwtToken, err := jwt4.Parse(accessToken, jwks.Keyfunc)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid token: " + err.Error()})
			return
		}

		fmt.Printf("Claims runtime type: %T\n", jwtToken.Claims)

		// Token is valid, you can use it
		c.JSON(200, gin.H{"message": "Token is valid", "claims": jwtToken.Claims})

		mapClaims, ok := jwtToken.Claims.(jwt4.MapClaims)
		if !ok {
			c.JSON(401, gin.H{"error": "Can not extract claims"})
			return
		}

		if !mapClaims.VerifyIssuer("https://udemy-tenant-tg.us.auth0.com/", true) {
			c.JSON(401, gin.H{"error": "Invalid issuer"})
			return
		}

		if !mapClaims.VerifyAudience("https://contacts.example.com", true) {
			c.JSON(401, gin.H{"error": "Required audience not found"})
			return
		}

		if !mapClaims.VerifyExpiresAt(time.Now().Unix(), true) {
			c.JSON(401, gin.H{"error": "Token has expired"})
			return
		}

		if !mapClaims.VerifyIssuedAt(time.Now().Unix(), true) {
			c.JSON(401, gin.H{"error": "Token issue time is in the future"})
			return
		}

		if !validateScope(mapClaims, "openid") {
			c.JSON(403, gin.H{"error": "Insufficient scope"})
			return
		}

		c.JSON(200, gin.H{"message": "You have access to the resource!"})
	})

	r.Run(":8081")
}

// hasScope checks "scope" (space-delimited), "scp" (array), or "permissions" (array).
func validateScope(claims jwt4.MapClaims, required string) bool {
	// 1) "scope": "read:contacts write:contacts"
	if s, ok := claims["scope"].(string); ok {
		log.Default().Printf("Scopes in token: %s", s)
		if slices.Contains(strings.Fields(s), required) {
			return true
		}
	}

	// 2) "scp": ["read:contacts", "write:contacts"]
	if arr, ok := claims["scp"].([]interface{}); ok {
		log.Default().Printf("Scopes in token: %v", arr)
		for _, v := range arr {
			if str, ok := v.(string); ok && str == required {
				return true
			}
		}
	}

	// 3) Auth0 RBAC "permissions": ["read:contacts"]
	if arr, ok := claims["permissions"].([]interface{}); ok {
		log.Default().Printf("Scopes in token: %v", arr)
		for _, v := range arr {
			if str, ok := v.(string); ok && str == required {
				return true
			}
		}
	}

	return false
}
