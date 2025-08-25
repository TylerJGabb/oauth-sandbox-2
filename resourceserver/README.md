# JWT-Protected Resource Server (Go + Gin)

This is a simple Go server that demonstrates how to protect resources with **JWT authentication** using [Gin](https://github.com/gin-gonic/gin), [golang-jwt](https://github.com/golang-jwt/jwt), and [keyfunc](https://github.com/MicahParks/keyfunc).

## Features

* Serves a protected resource at `/resource`.
* Uses Auth0’s JWKS endpoint to validate incoming JWTs.
* Automatically refreshes JWKS when unknown keys are encountered.
* Validates required audience (`https://contacts.example.com`).
* Returns JWT claims when valid.

## Requirements

* Go 1.20+
* An Auth0 tenant (or other OIDC provider with a JWKS endpoint).
* A valid JWT with the correct audience.

## Running

Clone and run the service:

```bash
go run main.go
```

The server starts on port **8081**.

## Usage

Send a request to the protected resource: (hint, you can get an access token from the [SPA](../spa-with-dynamic-backend/README.md))

```bash
curl -H "Authorization: Bearer <ACCESS_TOKEN>" http://localhost:8081/resource
```

### Responses

* **200 OK**: Token is valid, and audience matches.
* **401 Unauthorized**: Missing/invalid token, or audience check failed.

## Future Improvements

* Dynamically detect the token issuer (`iss`) from the JWT.
* Cache JWKS per issuer to support multiple issuers efficiently.
