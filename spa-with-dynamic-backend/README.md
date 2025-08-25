TL;DR:
1. run `docker compose up` in the root of the repo
2. run `go run ./spa-with-dynamic-backend` in the root of the repo
3. open `http://localhost:8080` in your browser
4. click "Start Login" button
  
This is an example of an SPA (Single Page Application) with a dynamic backend that uses oauth to facilitate the aquisition of access tokens from auth0. A sequence diagram illustrating the flow of authentication and token acquisition can be found below.

```mermaid
sequenceDiagram
  participant Browser as Browser (SPA via localhost:8080)
  participant SPA as SPA (Frontend Script)
  participant Auth0 as Auth0
  participant Backend as Backend Server
  participant Redis as Redis Store

  Browser ->> Backend: GET / (request index.html)
  Backend -->> Browser: index.html
  Browser ->> SPA: Load index.html (start SPA)
  Note right of Browser: User clicks "login"
  SPA ->> Auth0: Redirect to Auth0 login page
  Auth0 ->> Browser: Show login form
  Browser ->> Auth0: User submits credentials
  Auth0 ->> SPA: Redirect with authorization code
  SPA ->> Backend: REDIRECT /oauth-exchange<br>with authorization code
  Backend ->> Auth0: Exchange code for access token
  Auth0 -->> Backend: Access token
  Backend ->> Redis: Store access token
  Backend -->> Browser: Set opaque session cookie
  Note over SPA, Browser: SPA stores opaque session in memory
  Note over SPA, Browser: All future requests<br/>authenticated via session cookie
```
