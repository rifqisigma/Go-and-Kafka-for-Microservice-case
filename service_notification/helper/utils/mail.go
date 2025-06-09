package utils

import (
	"fmt"
	"os"
	"service_notification/dto"

	"gopkg.in/gomail.v2"
)

func SendEmail(input *dto.SendEmail) {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", os.Getenv("EMAIL_SENDER"))
	mailer.SetHeader("To", input.ToEmail)

	header := fmt.Sprintf("Shop Notification:%s", input.Header)
	mailer.SetHeader("Subject", header)
	mailer.SetBody("text/html", input.Desc)
	dialer := gomail.NewDialer("smtp.gmail.com", 587, os.Getenv("EMAIL_SENDER"), os.Getenv("APP_PASSWORD"))

	if err := dialer.DialAndSend(mailer); err != nil {
		fmt.Println("Error sending email:", err)
	}
}
