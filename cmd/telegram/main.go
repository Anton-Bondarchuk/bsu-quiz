package main

import (
	"bsu-quiz/internal/app/telegram"
	"context"
)

func main() {
	app, closeFunc := telegram.NewAppTelegram()
	defer func() {
		err := closeFunc()
		if err != nil {
			panic(err)
		}
	}()

	telegram.Start(context.Background(), app)
}
