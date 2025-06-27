package main

import (
	"bsu-quiz/quiz/internal/app/admin"
	"context"
)

func main() {
	// NOTE: handle context in this place 
	// or in the level above and give context with Timeout 
	// for correct handle locking database or someone else strucure 
	app := admin.NewAdminApp()

	admin.Start(context.Background(), app)
}