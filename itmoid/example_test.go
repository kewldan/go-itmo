package itmoid_test

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/kewldan/go-itmo/itmoid"
)

// Accounts with two-factor authentication supply the one-time code on demand.
func ExampleAuthenticator_Login_otp() {
	auth := itmoid.New()
	tok, err := auth.Login(context.Background(), itmoid.MyITMO, itmoid.Credentials{
		Username: os.Getenv("ITMO_LOGIN"),
		Password: os.Getenv("ITMO_PASSWORD"),
		OTP: func(context.Context) (string, error) {
			fmt.Print("one-time code: ")
			return bufio.NewReader(os.Stdin).ReadString('\n')
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	claims, _ := itmoid.IDClaims(tok)
	fmt.Println("signed in as", claims.Name)
}

// Let the user sign in on the real ITMO.ID page (browser or WebView) and
// catch the redirect to the callback; the password never reaches the app.
func ExampleParseCallback() {
	auth := itmoid.New()
	state, verifier := itmoid.NewState(), "" // use oauth2.GenerateVerifier() for PKCE apps
	fmt.Println("open:", auth.AuthCodeURL(itmoid.BARS, state, verifier))

	redirected := "https://bars.itmo.ru/rest/login?state=...&code=..." // caught by the WebView
	code, err := itmoid.ParseCallback(itmoid.BARS, auth.Issuer(), redirected, state)
	if err != nil {
		log.Fatal(err)
	}
	_ = code // pass to bars.Client.Login
}
