# Form3 Take Home Exercise

## Example

```go
service := form3.NewClient("http://localhost:8080")

orgs, err := service.List(form3.PageNumber(0), form3.PageSize(25))
if err != nil {
	log.Fatalf("could not list organizations: %v", err)
}

log.Printf("We have %d organizations\n", len(orgs))

err = service.Delete("f4f3fa9f-261c-458e-b032-9bfa45aa091c", 0)
if err != nil {
	log.Fatalf("could not delete organization: %v", err)
}

org, err = service.Fetch("f4f3fa9f-261c-458e-b032-9bfa45aa091c")
if err != nil {
	log.Fatalf("could not fetch organization: %v", err)
}

log.Printf("We got organization with ID: %s\n", org.ID)
```