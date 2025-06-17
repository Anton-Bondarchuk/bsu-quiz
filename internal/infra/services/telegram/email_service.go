package tgservices

import (
	"bsu-quiz/internal/infra/clients"
	"fmt"
	"time"
)

type EmailClienter interface {
	Send(login, subject string, data map[string]any) error
}

type EmailService struct {
	client EmailClienter
}

func NewEmailService(client *clients.EmailClient) *EmailService {
	return &EmailService{
		client: client,
	}
}

func (s *EmailService) Send(login, subject, code string, expiresAt time.Time) error {
	data := map[string]any{
		"Login":     login,
		"Code":      code,
		"ExpiresIn": fmt.Sprintf("%.0f minutes", time.Until(expiresAt).Minutes()),
	}

	return s.client.Send(login, subject, data)
}
