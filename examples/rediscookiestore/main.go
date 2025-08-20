package main

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

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
		MaxAge:   5,
		HttpOnly: true,                 // not readable by JS
		Secure:   false,                // if this is deployed behind TLS, set this to true
		SameSite: http.SameSiteLaxMode, // adjust as needed
		// (?) what about wildcards?
		// Domain:  "your.domain", // set if you need a specific domain
	})

	r.Use(sessions.Sessions("session-storage", store))

	r.GET("/login", func(ctx *gin.Context) {
		name := ctx.Query("name")
		session := sessions.Default(ctx)
		session.Set("user", name)
		session.Save()
		ctx.JSON(http.StatusOK, gin.H{"message": "Logged in"})
	})

	r.GET("/profile", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		user := session.Get("user")
		if user == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized, session invalid or expired"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"user": user})
	})

	r.GET("/logout", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		session.Clear()
		session.Save()
		ctx.JSON(http.StatusOK, gin.H{"message": "Logged out"})
	})

	r.Run(":8080")

}
