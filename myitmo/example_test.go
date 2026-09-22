package myitmo_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2"

	"github.com/kewldan/go-itmo/itmoid"
	"github.com/kewldan/go-itmo/myitmo"
)

// Sign in with a password once, then keep only the refresh token.
func Example() {
	ctx := context.Background()
	auth := itmoid.New()

	tok, err := auth.Login(ctx, itmoid.MyITMO, itmoid.Credentials{
		Username: os.Getenv("ITMO_LOGIN"),
		Password: os.Getenv("ITMO_PASSWORD"),
	})
	if err != nil {
		log.Fatal(err)
	}

	// ITMO.ID rotates refresh tokens: persist every token it hands out.
	save := func(t *oauth2.Token) { _ = os.WriteFile(".refresh-token", []byte(t.RefreshToken), 0o600) }
	save(tok)

	client := myitmo.New(auth.TokenSource(ctx, itmoid.MyITMO, tok, save))

	today := myitmo.Today()
	days, err := client.Schedule.Personal(ctx, today, today.AddDays(7))
	if err != nil {
		log.Fatal(err)
	}
	for _, day := range days {
		for _, l := range day.Lessons {
			fmt.Println(day.Date, l.TimeStart, l.Subject, l.Room)
		}
	}
}

// Resume from a stored refresh token.
func ExampleNew_refreshToken() {
	ctx := context.Background()
	auth := itmoid.New()
	stored, _ := os.ReadFile(".refresh-token")

	ts := auth.FromRefreshToken(ctx, itmoid.MyITMO, string(stored), func(t *oauth2.Token) {
		_ = os.WriteFile(".refresh-token", []byte(t.RefreshToken), 0o600)
	})
	client := myitmo.New(ts)

	specs, err := client.RecordBook.Specializations(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range specs {
		for _, sem := range s.Semesters {
			entries, err := client.RecordBook.Entries(ctx, s.MainPlan, sem.Semester)
			if err != nil {
				log.Fatal(err)
			}
			for _, e := range entries {
				fmt.Println(sem.Semester, e.Name, e.Rate)
			}
		}
	}
}

// Server-side failures carry the MyITMO error code and message.
func ExampleError() {
	var client *myitmo.Client // configured elsewhere
	_, err := client.RecordBook.Specializations(context.Background())
	var apiErr *myitmo.Error
	switch {
	case err == nil:
	case errors.As(err, &apiErr) && apiErr.IsForbidden():
		fmt.Println("section is not available to this account")
	default:
		fmt.Println(myitmo.ErrorCode(err), err)
	}
}
