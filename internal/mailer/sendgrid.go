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

func (m SendGridMailer) Send(templateFile, username, email string, data any, isSandbox bool) error {
	// Sender and receiver email
	from := mail.NewEmail(FromName, m.fromEmail)
	to := mail.NewEmail(username, email)

	// template parsing and building
	tmpl, err := template.ParseFS(FS, "template/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return err
	}

	body := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(body, "body", data); err != nil {
		return err
	}

	message := mail.
		NewSingleEmail(from, subject.String(), to, "", body.String())

	// Set sandbox mode for testing
	message.MailSettings = &mail.MailSettings{
		SandboxMode: &mail.Setting{Enable: &isSandbox},
	}

	for i := 0; i < maxRetries; i++ {
		response, err := m.client.Send(message)
		if err != nil {
			log.Printf("Attempt %d: failed to send email: %v", i+1, err)
			log.Printf("Error response: %v", err)

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
			return fmt.Errorf("sendgrid refused the email to %s (status %d): %s", email, response.StatusCode, response.Body)
		}

		// Email send successful
		log.Printf("Email sent successfully to %s. Status Code: %d", email, response.StatusCode)
		return nil
	}

	return fmt.Errorf("failed to send email to %s after %d attempts", email, maxRetries)
}
