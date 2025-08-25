A collection of OAuth 2.0 and OpenID Connect examples using Auth0.

1. Single Page Application (SPA) with Dynamic Backend [spa-with-dynamic-backend/](spa-with-dynamic-backend/)
   1. note: you need to run redis via `docker compose up` which starts a containerized redis instance locally
2. Redis backend session cookies with redis and gin [examples/rediscookiestore](examples/rediscookiestore)
3. Reading a session from redis (this is more of a tool) [examples/readsession](examples/readsession)
4. Validating tokens as the resource server [resourceserver](resourceserver)