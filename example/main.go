package main

import (
	"log"

	"github.com/nclandrei/form3"
)

func main() {
	service := form3.NewClient("http://localhost:8080")

	orgs, err := service.List()
	if err != nil {
		log.Fatalf("could not list organizations: %v", err)
	}

	log.Printf("We have %d organizations\n", len(orgs))

	err = service.Delete(orgs[0].ID, orgs[0].Version)
	if err != nil {
		log.Fatalf("could not delete organization: %v", err)
	}

	org, err := service.Fetch(orgs[1].ID)
	if err != nil {
		log.Fatalf("could not delete organization: %v", err)
	}

	log.Printf("Organization %s successfully fetched!\n", org.ID.String())
}
