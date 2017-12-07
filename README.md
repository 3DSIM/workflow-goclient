[![GoDoc](https://godoc.org/github.com/3DSIM/workflow-goclient?status.svg)](https://godoc.org/github.com/3DSIM/workflow-goclient)

# workflow-goclient
Go client for interacting with the 3DSIM workflow api.

## Technical Specifications

### Platforms Supported
MacOS, Windows, and Linux

## Background Info
We use https://goswagger.io to generate our Go APIs and clients.  This allows
us to build our APIs in a "design first" manner.

First we create a `swagger.yaml` file that defines the API.  Then we run a command
to generate the server code.

Additionally, this allows us to automatically generate client code.  The code in this
directory was all generated using the `go-swagger` tools.


## Organization
* `workflow` - the client package that adds convenience methods for common workflow operations
* `genclient` - the generated client code
* `models` - the generated models

## Regenerating code
First install the swagger generator.  Currently we are using release 0.11.0 of https://github.com/go-swagger/go-swagger.

For mac users:
* brew tap go-swagger/go-swagger
* brew install go-swagger

For windows users:
* See https://github.com/go-swagger/go-swagger for options

The code generator needs a specification file.  The specification for the workflow API is stored in `github.com/3dsim/workflow/swagger.yaml`.  Assuming that project
is cloned as a sibling project, the command to run to generate new client code is:
```
swagger generate client -A WorkflowAPI -f ../workflow/swagger.yaml --client-package genclient
```

* Generate fakes using counterfeiter
```
go get github.com/maxbrunsfeld/counterfeiter
```
From inside package folder
```
go generate
```

* Generate mocks using https://github.com/vektra/mockery

```
go get github.com/vektra/mockery/.../
$GOPATH/bin/mockery -name <interface name> -recursive
```

If you need to generate mocks in the same package to avoid circular dependencies use
```
$GOPATH/bin/mockery -name <interface name> -recursive -inpkg
```

## Using the client
TODO

## Contributors
* Tim Sublette
* Ryan Walls
* Chad Queen
* Pete Krull
* Alex Drinkwater

## Original release
December 2017