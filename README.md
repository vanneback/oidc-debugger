This repo is to make debugging OIDC / oauth2 flow easier. It prints all the tokens
and claims on a webpage hosten on localhost:5000.

## Usage
* Set the env variables from example env in your shell. 
* Configure the oidc-provider with the same client_id and client_secret. And set 
redirect urls to "http://localhost:5000/auth/oidc/callback".
* `go mod download`
* `./main.go`