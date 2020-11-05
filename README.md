# Form3 Take Home Exercise (Name: Andrei-Mihai Nicolae)

## Overview

This is a client library written in Go that interacts with a fake Form3 API. It supports the following operations **exclusively on Organisation Accounts**:

* List - lists organisation accounts (paging optional)
* Fetch - retrieves a single organisation account if it exists
* Delete - deletes an organisation account given its ID
* Create - creates a new organisation account

## Running the tests

For simply running the tests against the provided fake API, simply run the following command in a terminal:

```bash
$ make run-tests
```

It works, of course, using simply ```docker-compose up```, but the make command will clean the cache and rebuild it from scratch, automatically exit once tests are run, as well as show output only from the client container, so it's easier for the reader to see test results.

## Example of using the client library

```go
service := form3.NewClient("http://localhost:8080")

// get organisation accounts stored in Form3 using paging functionality
orgs, err := service.List(form3.PageNumber(0), form3.PageSize(25))

// remove an organisation account with the ID and version below
err = service.Delete("f4f3fa9f-261c-458e-b032-9bfa45aa091c", 0)

// retrieve a single organisation using the ID below
org, err = service.Fetch("f4f3fa9f-261c-458e-b032-9bfa45aa091c")

// creates an organisation account insidde Form3
org, err = service.Create(form3.OrganisationAccount{...})
```

## Technical Decisions

In this section I will describe in more detail multiple decisions I made throughout the implementation process.

### Testing

The test coverage is ~85% and it's not bigger because the only paths that weren't tested were mainly if the ```json``` package fails to encode/decode. Apart from that, absolutely all paths and possible responses from the API are tested.

I tried to use as few dependencies as possible to keep the project small. However, [testify](https://github.com/stretchr/testify) is amazing for testing and I use it all the time for testing in Go projects, thus I went with it :)

I also tried to use best practices for testing:

* use table driven testing
* use bulletproof framework for testing
* follow all paths, even unhappy ones
* reach high coverage
* use idiomatic testdata folder for having test data to use when running tests

### Models

I have added only 2 models in the app: ```OrganisationAccount``` and ```OrganisationAccountAttributes```. 

Both of them are usable constructing them directly as a struct, thus not providing any constructors. 

The main reason I have not gone with constructors is because there are a lot of fields, which would have made the constructor horribly to use. Also, another option would have been to use the Builder design pattern, but that is not so idiomatic in Go (as in Java for example), thus I have let the users construct the structs as they please.

Another decision that I made regarding models was related to **validation**. I know that a lot of the fields have specific requirements and I could have validated them when constructing the object, but the API already performs those validations, thus when returning the result to the caller, I simply wrap the validation errors from the API into new Go errors. If I were to implement validations, I consider it as a redundant code duplication.

### Service

Quite straightforward - 4 different methods for supporting the 4 methods we're interested in. I also added in a rate limiter that wraps all calls so that the client library can avoid unavailable service or other similar errors.

One important note - I willingly avoided returning links to the caller due to my assumption that users are interested only in organisation accounts. That would have been just a simple couple of lines addition! :)
