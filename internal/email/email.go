package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

func SendEmail(to string, message []byte) error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %v", err.Error())
	}

	from := os.Getenv("EMAIL_FROM")
	user := os.Getenv("EMAIL_SERVER_USER")
	password := os.Getenv("EMAIL_SERVER_PASSWORD")
	smtpHost := os.Getenv("EMAIL_SERVER_HOST")
	smtpPort := os.Getenv("EMAIL_SERVER_PORT")

	conn, err := smtp.Dial(smtpHost + ":" + smtpPort)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err.Error())
	}
	defer conn.Close()

	// Send EHLO
	err = conn.Hello("localhost")
	if err != nil {
		return fmt.Errorf("EHLO failed: %v", err.Error())
	}

	// Start TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false, // Should be true only for testing with self-signed certs
		ServerName:         smtpHost,
	}

	err = conn.StartTLS(tlsConfig)
	if err != nil {
		return fmt.Errorf("STARTTLS failed: %v", err.Error())
	}

	// Re-send EHLO after TLS is established
	err = conn.Hello("localhost")
	if err != nil {
		return fmt.Errorf("EHLO after STARTTLS failed: %v", err.Error())
	}

	// Authenticate
	auth := smtp.PlainAuth("", user, password, smtpHost)
	err = conn.Auth(auth)
	if err != nil {
		return fmt.Errorf("authentication failed: %v", err.Error())
	}

	// Send email
	err = conn.Mail(from)
	if err != nil {
		return fmt.Errorf("MAIL FROM failed: %v", err.Error())
	}

	err = conn.Rcpt(to)
	if err != nil {
		return fmt.Errorf("RCPT TO failed: %v", err.Error())
	}

	wc, err := conn.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %v", err.Error())
	}

	_, err = wc.Write(message)
	if err != nil {
		return fmt.Errorf("writing message failed: %v", err.Error())
	}

	err = wc.Close()
	if err != nil {
		return fmt.Errorf("closing message failed: %v", err.Error())
	}

	// Quit session
	err = conn.Quit()
	if err != nil {
		return fmt.Errorf("QUIT failed: %v", err.Error())
	}

	fmt.Println("Email sent successfully!")

	return nil
}
