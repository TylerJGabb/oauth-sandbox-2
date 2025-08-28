package main

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type CustomClaims struct {
	jwt.RegisteredClaims
	Permissions []string `json:"permissions"`
}

func (c *CustomClaims) Validate(permissions []string) error {
	if !c.VerifyIssuer("https://udemy-tenant-tg.us.auth0.com/", true) {
		return fmt.Errorf("invalid issuer")
	}
	if !c.VerifyAudience("api:my-test-api", true) {
		return fmt.Errorf("invalid audience")
	}
	if !c.VerifyExpiresAt(time.Now(), true) {
		return fmt.Errorf("token has expired")
	}
	if !c.VerifyIssuedAt(time.Now(), true) {
		return fmt.Errorf("token is not yet valid")
	}
	for _, p := range permissions {
		if !slices.Contains(c.Permissions, p) {
			return fmt.Errorf("insufficient permissions: %+v", c.Permissions)
		}
	}
	return nil
}

func requiresPermissions(jwks *keyfunc.JWKS, permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the token from the Authorization header
		bearerToken := c.Request.Header.Get("Authorization")
		if bearerToken == "" {
			c.JSON(401, gin.H{"error": "Missing Authorization Header"})
			c.Abort()
			return
		}

		accessToken := strings.TrimPrefix(bearerToken, "Bearer ")

		customClaims := CustomClaims{}
		_, err := jwt.ParseWithClaims(accessToken, &customClaims, jwks.Keyfunc)
		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		if err := customClaims.Validate(permissions); err != nil {
			c.JSON(401, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Next()
	}
}

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

	r.GET("/resource", requiresPermissions(jwks, "test:read"), func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "You have access to the resource!"})
	})

	r.Run(":8081")
}
