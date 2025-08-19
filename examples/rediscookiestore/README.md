this program starts a webserver that uses cookies to persist session data in Redis

`go run ./examples/rediscookiestore`

- A call to `localhost:8080/login?name=${YOUR_NAME}` will log in a user and create a session. **The session cookie is opaque, with its contents stored in Redis.**
- A call to `localhost:8080/profile` will return the current user's profile information.
- A call to `localhost:8080/logout` will log out the user and destroy the session.
