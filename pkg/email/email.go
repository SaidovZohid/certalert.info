package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"github.com/SaidovZohid/certalert.info/config"
)

type SendEmailRequest struct {
	To      []string
	Type    string
	Body    map[string]string
	Subject string
}

const (
	VerificationEmail   = "verification_email"
	ForgotPasswordEmail = "forgot_password_email"
	ChangeEmail = "change_email"
)

func SendEmail(cfg *config.Config, req *SendEmailRequest) error {
	from := cfg.Smtp.Sender
	to := req.To

	password := cfg.Smtp.Password

	var body bytes.Buffer

	templatePath := getTemplatePath(req.Type)
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}
	t.Execute(&body, req.Body)

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := fmt.Sprintf("Subject: %s\n", req.Subject)
	msg := []byte(subject + mime + body.String())

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	err = smtp.SendMail("smtp.gmail.com:587", auth, from, to, msg)
	if err != nil {
		return err
	}
	return nil
}

func getTemplatePath(emailType string) string {
	switch emailType {
	case VerificationEmail:
		return "./templates/verification_email.html"
	case ForgotPasswordEmail:
		return "./templates/forgot_password_email.html"
	case ChangeEmail:
		return "./templates/change_email.html"
	}
	return ""
}
