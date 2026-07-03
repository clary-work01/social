package mailer

import "embed"

const (
	FromName            = "Chainflow"
	maxRetries          = 3
	UserWelcomeTemplate = "user_invitation.tmpl"
)

//go:embed templates
var FS embed.FS

type Client interface {
	// isSandbox:開發環境不會真的發email
	Send(templateFiles, username, email string, data any, isSandbox bool) (int, error)
}
