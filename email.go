package main

import (
	"fmt"

	"gopkg.in/mail.v2"
)

// Function to send a welcome email
func sendEmail(to, name, subject, country, time, phone string) error {
	m := mail.NewMessage()
	m.SetHeader("From", "FxOpus <contact@fxopus.com>") // Ensure emailConfig is properly initialized
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)

	// HTML email body
	//var body string
	body := fmt.Sprintf(`Hi %s,<br>
We have received your request and our team will get in touch with you very soon. Thank you for considering FxOpus.<br><br>
<strong>Phone Number:</strong> %s<br>
<strong>Preferred Time:</strong> %s<br>
<strong>Country:</strong> %s`,
		name, phone, time, country)

	// Set the email body
	m.SetBody("text/html", body)
	d := mail.NewDialer("smtp.zoho.in", 587, "contact@fxopus.com", "ZbkV7YRmUQy7")

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func sendLossRecoveryEmail(name, to, country, whatsapp string, accountSize, pastLoss string, hasRunningTrades bool) error {
	m := mail.NewMessage()
	m.SetHeader("From", "FxOpus <contact@fxopus.com>") // Ensure emailConfig is properly initialized
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Loss recovery program")
	runningTrades := "False"
	if hasRunningTrades {
		runningTrades = "True"
	}

	// HTML email body
	//var body string
	body := fmt.Sprintf(`Hi %s,<br>
We have received your request and our team will get in touch with you very soon. Thank you for considering FxOpus.<br><br>
<strong>Phone Number:</strong> %s<br>
<strong>Account Size:</strong> %s<br>
<strong>Loss size:</strong> %s<br>
<strong>Country:</strong> %s<br>
Running trades present: %s`,
		name, whatsapp, accountSize, pastLoss, country, runningTrades)

	// Set the email body
	m.SetBody("text/html", body)
	d := mail.NewDialer("smtp.zoho.in", 587, "contact@fxopus.com", "ZbkV7YRmUQy7")

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
