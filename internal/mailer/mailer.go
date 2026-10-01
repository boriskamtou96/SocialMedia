package mailer

import "embed"

const (
	FromName               = "GopherSocial"
	UserInvitationTemplate = "user_invitation.tmpl"
	maxRetries             = 3
)

//go:embed "template"
var FS embed.FS

type Client interface {
	Send(templateFile, username, email string, data any, isSandbox bool) error
}
