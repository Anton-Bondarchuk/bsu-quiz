package main

import (
	"bsu-quiz/internal/config"
	"bsu-quiz/internal/infra/clients"
	tgservices "bsu-quiz/internal/infra/services/telegram"
	"log"
	"time"
)

func main() {
	cfg := config.MustLoad()

	emailClient := clients.NewEmailClient(cfg.EmailConfig)
	emailService := tgservices.NewEmailService(emailClient)


	otp := "1234"
	expiresAt := time.Now().Add(time.Minute * 5) // OTP valid for 5 minutes
	login := "rct.bondarchAS"

	if err := emailService.Send(login, "Your Verification Code", otp, expiresAt); err != nil {
		log.Printf("Failed to sentd verification email: %v", err)
	} else {
		log.Printf("Verification email sent to %s", login)
	}
}