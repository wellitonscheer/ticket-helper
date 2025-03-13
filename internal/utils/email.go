package utils

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/wellitonscheer/ticket-helper/internal/config"
	gomail "gopkg.in/mail.v2"
)

func SendEmail(emailConf config.EmailConfig, to, subject, message string) error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %v", err.Error())
	}

	sendMessage := gomail.NewMessage()

	sendMessage.SetHeader("From", emailConf.From)
	sendMessage.SetHeader("To", to)
	sendMessage.SetHeader("Subject", subject)

	sendMessage.SetBody("text/plain", message)

	dialer := gomail.NewDialer(emailConf.Host, emailConf.Port, emailConf.User, emailConf.Password)

	if err := dialer.DialAndSend(sendMessage); err != nil {
		return fmt.Errorf("failed to send email: %v", err.Error())
	}

	return nil
}
