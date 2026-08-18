package main

import (
	"os"

	"d5/internal/handler"
	"d5/internal/repository"
)

func main() {
	users := repository.NewUserRepository()
	app := handler.NewUserHandler(users, os.Stdin, os.Stdout)
	app.Run()
}
