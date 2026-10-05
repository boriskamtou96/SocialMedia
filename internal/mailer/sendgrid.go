package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
}

func NewSendgrid(apiKey, fromEmail string) *SendGridMailer {
	client := sendgrid.NewSendClient(apiKey)
	return &SendGridMailer{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		client:    client,
	}
}

func (m SendGridMailer) Send(templateFile, username, email string, data any, isSandbox bool) (int, error) {
	// Sender and receiver email
	from := mail.NewEmail(FromName, m.fromEmail)
	to := mail.NewEmail(username, email)

	// template parsing and building
	tmpl, err := template.ParseFS(FS, "template/"+templateFile)
	if err != nil {
		return -1, err
	}

	subject := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return -1, err
	}

	body := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(body, "body", data); err != nil {
		return -1, err
	}

	message := mail.
		NewSingleEmail(from, subject.String(), to, "", body.String())

	// Set sandbox mode for testing
	message.MailSettings = &mail.MailSettings{
		SandboxMode: &mail.Setting{Enable: &isSandbox},
	}

	var retryError error
	for i := 0; i < maxRetries; i++ {
		response, retryError := m.client.Send(message)
		if retryError != nil {
			// exponential backoff before retrying

			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		// SendGrid reports refusals (invalid API key, unverified sender...) as a response, not as an error
		switch {
		case response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500:
			log.Printf("Attempt %d: sendgrid unavailable (status %d), retrying", i+1, response.StatusCode)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		case response.StatusCode >= 300:
			return -1, fmt.Errorf("sendgrid refused the email to %s (status %d): %s", email, response.StatusCode, response.Body)
		}

		// Email send successful
		return response.StatusCode, nil
	}

	return -1, fmt.Errorf("failed to send email to %s after %d attempts, error: %v", email, maxRetries, retryError)
}
