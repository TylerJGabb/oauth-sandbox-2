package main

import (
	"fmt"
	"log"
	"strings"

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

		if !mapClaims.VerifyAudience("https://contacts.example.com", true) {
			c.JSON(401, gin.H{"error": "Required audience not found"})
			return
		}

		c.JSON(200, gin.H{"message": "You have access to the resource!"})
	})

	r.Run(":8081")
}
