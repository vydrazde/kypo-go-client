# kypo-go-client
A [KYPO CRP](https://docs.crp.kypo.muni.cz/) client library written in Go.

## Supported API calls:
- Login to CSIRT-MU Dummy OIDC and Keycloak
- Sandbox Definition - Get, Create, Delete
- Sandbox Pool - Get, Create, Delete, Cleanup
- Sandbox Allocation Unit - Get, CreateAllocation, CreateAllocationAwait, CancelAllocation, CreateCleanup, CreateCleanupAwait, GetAllocationOutput
- Training Definition - Get, Create, Delete
- Training Definition Adaptive - Get, Create, Delete

## Usage
```go
import "github.com/vydrazde/kypo-go-client"
```

Create a client with username and password:
```go
client, err := kypo.NewClient("https://your.kypo.ex", "KYPO-Client", "username", "password")
if err != nil {
    log.Fatalf("Failed to create KYPO client: %v", err)
}
```

Use the client to create a sandbox definition:
```go
sandboxDefinition, err := client.CreateSandboxDefinition(context.Background(), 
	"git@gitlab.ics.muni.cz:kypo-library/content/kypo-library-demo-training.git", "master")

if err != nil {
    log.Fatalf("Failed to create sandbox definition: %v", err)
}
```

