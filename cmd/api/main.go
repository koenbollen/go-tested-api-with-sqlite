package main

import (
	"context"

	"github.com/koenbollen/go-tested-api-with-sqlite/internal"
	"github.com/koenbollen/go-tested-api-with-sqlite/internal/routes"
)

func main() {
	internal.Main(context.Background(), "api", routes.Redirections)
}
