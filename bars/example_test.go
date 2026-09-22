package bars_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/kewldan/go-itmo/bars"
	"github.com/kewldan/go-itmo/itmoid"
)

// Read the own journal for every discipline of a period.
func Example() {
	ctx := context.Background()
	auth := itmoid.New()
	// The authenticator doubles as the renewal source: when the ~30 minute
	// session expires, the client gets a new code through the SSO cookies.
	client := bars.New(bars.WithAuthenticator(auth))

	if err := client.LoginPassword(ctx, itmoid.Credentials{
		Username: os.Getenv("ITMO_LOGIN"),
		Password: os.Getenv("ITMO_PASSWORD"),
	}); err != nil {
		log.Fatal(err)
	}

	err := client.WithPeriod(ctx, "2025/2026", bars.Spring, func(ctx context.Context) error {
		disciplines, err := client.Disciplines(ctx, true)
		if err != nil {
			return err
		}
		flows, err := client.GroupsAndFlows(ctx, 0)
		if err != nil {
			return err
		}
		for _, d := range disciplines {
			for _, plan := range d.CheckpointPlanIDs {
				for _, f := range flows {
					if !slices.Contains(f.CheckpointPlanIDs, plan) {
						continue
					}
					j, err := client.StudentJournal(ctx, plan, f.Type, f.Identifier)
					if err != nil {
						return err
					}
					for _, s := range j.Students {
						if s.Marks.HasAnyMark() && s.Marks.Total != nil {
							fmt.Printf("%s: %.1f\n", d.Name, *s.Marks.Total)
						}
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
