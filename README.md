# go-jackett

It is non-official Golang SDK for [Jackett](https://github.com/Jackett/Jackett).

Example usage:

```golang
package main

import (
	"log"
	"context"
	"github.com/webtor-io/go-jackett"
)

func main() {
    ctx := context.Background()
    j := jackett.NewJackett(&jackett.Settings{
        ApiURL: "YOUR_API_URL",
        ApiKey: "YOUR_API_KEY",
    })
    resp, err := j.Fetch(ctx, &jackett.FetchRequest{
        Categories: []uint{7000},
        Query:      "Crime and Punishment",
    })
    if err != nil {
        panic(err)
    }
    for _, r := range resp.Results {
        log.Printf("%+v", r)
    }

    // Adding a new indexer
    err = j.AddIndexer(ctx, "new-indexer")
    if err != nil {
        panic(err)
    }

    // Removing an existing indexer
    err = j.RemoveIndexer(ctx, "existing-indexer")
    if err != nil {
        panic(err)
    }

    // Updating an existing indexer
    settings := map[string]interface{}{
        "setting1": "value1",
        "setting2": "value2",
    }
    err = j.UpdateIndexer(ctx, "existing-indexer", settings)
    if err != nil {
        panic(err)
    }
}
```

As ApiUrl just use root url of your Jackett instance. ApiKey could be found at the top of Jackett UI.

It is also possible to get Jackett credentials from environment variables `JACKETT_API_URL` and `JACKETT_API_KEY`.
In this case just provide empty settings like so:

```golang
j := jackett.NewJackett(&jackett.Settings{})
```

## New Methods

### AddIndexer

Adds a new indexer to the Jackett instance.

```golang
func (j *Jackett) AddIndexer(ctx context.Context, indexerID string) error
```

### RemoveIndexer

Removes an existing indexer from the Jackett instance.

```golang
func (j *Jackett) RemoveIndexer(ctx context.Context, indexerID string) error
```

### UpdateIndexer

Updates the settings of an existing indexer in the Jackett instance.

```golang
func (j *Jackett) UpdateIndexer(ctx context.Context, indexerID string, settings map[string]interface{}) error
```
