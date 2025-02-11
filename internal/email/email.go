package email

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	gomail "gopkg.in/mail.v2"
)

func SendEmail(to, subject, message string) error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %v", err.Error())
	}

	from := os.Getenv("EMAIL_FROM")
	user := os.Getenv("EMAIL_SERVER_USER")
	password := os.Getenv("EMAIL_SERVER_PASSWORD")
	smtpHost := os.Getenv("EMAIL_SERVER_HOST")
	smtpPort, err := strconv.Atoi(os.Getenv("EMAIL_SERVER_PORT"))
	if err != nil {
		return fmt.Errorf("failed to convert port: %v", err.Error())
	}

	sendMessage := gomail.NewMessage()

	sendMessage.SetHeader("From", from)
	sendMessage.SetHeader("To", to)
	sendMessage.SetHeader("Subject", subject)

	sendMessage.SetBody("text/plain", message)

	dialer := gomail.NewDialer(smtpHost, smtpPort, user, password)

	if err := dialer.DialAndSend(sendMessage); err != nil {
		return fmt.Errorf("failed to send email: %v", err.Error())
	}

	return nil
}
