package main

import (
	"context"
	"fmt"
	"os"

	"d10-go-cli-application/internal/handler"
	"d10-go-cli-application/internal/logger"
	"d10-go-cli-application/internal/repository"
	"d10-go-cli-application/internal/service"
)

func main() {
	// Create the composition root
	repo := repository.NewMemoryUserRepository()

	log, err := logger.NewAsyncLogger(os.Stdout, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := log.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to close logger: %v\n", err)
		}
	}()

	svc := service.NewUserService(repo, log)
	app := handler.NewUserHandler(svc, os.Stdin, os.Stdout)

	ctx := context.Background()
	app.Run(ctx)
}
