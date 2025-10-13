package services

import (
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
)

type EmailService struct {
	client *resend.Client
}

func NewEmailService() *EmailService {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		// For development, we'll log instead of sending emails
		return &EmailService{client: nil}
	}
	
	client := resend.NewClient(apiKey)
	return &EmailService{client: client}
}

func (e *EmailService) SendVerificationEmail(email, firstName, verificationToken string) error {
	if e.client == nil {
		// Development mode - just log
		fmt.Printf("📧 [DEV] Verification email for %s:\n", email)
		fmt.Printf("   Token: %s\n", verificationToken)
		fmt.Printf("   URL: http://localhost:5173/verify-email?token=%s\n", verificationToken)
		return nil
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173" // fallback for development
	}
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", frontendURL, verificationToken)
	
	params := &resend.SendEmailRequest{
		From:    "Book Tracker <noreply@booktracker.rustyphillips.net>",
		To:      []string{email},
		Subject: "Verify your email address",
		Html: fmt.Sprintf(`
			<h1>Welcome to Book Tracker!</h1>
			<p>Hi %s,</p>
			<p>Thank you for registering! Please click the link below to verify your email address:</p>
			<p><a href="%s">Verify Email Address</a></p>
			<p>If the button doesn't work, copy and paste this URL into your browser:</p>
			<p>%s</p>
			<p>This link will expire in 24 hours.</p>
			<p>If you didn't create this account, you can safely ignore this email.</p>
		`, firstName, verificationURL, verificationURL),
	}

	_, err := e.client.Emails.Send(params)
	return err
}

func (e *EmailService) SendInvitationEmail(email, inviterName, childName, verificationToken string) error {
	if e.client == nil {
		// Development mode - just log
		fmt.Printf("📧 [DEV] Invitation email for %s:\n", email)
		fmt.Printf("   Inviter: %s\n", inviterName)
		fmt.Printf("   Child: %s\n", childName)
		fmt.Printf("   Token: %s\n", verificationToken)
		fmt.Printf("   URL: http://localhost:5173/accept-invitation?token=%s\n", verificationToken)
		return nil
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173" // fallback for development
	}
	acceptURL := fmt.Sprintf("%s/accept-invitation?token=%s", frontendURL, verificationToken)
	
	params := &resend.SendEmailRequest{
		From:    "Book Tracker <noreply@booktracker.rustyphillips.net>",
		To:      []string{email},
		Subject: fmt.Sprintf("%s has invited you to track %s's reading progress", inviterName, childName),
		Html: fmt.Sprintf(`
			<h1>You've been invited to Book Tracker!</h1>
			<p>%s has invited you to help track %s's reading progress.</p>
			<p>Click the link below to create your account and start tracking:</p>
			<p><a href="%s">Accept Invitation</a></p>
			<p>If the button doesn't work, copy and paste this URL into your browser:</p>
			<p>%s</p>
			<p>This invitation will expire in 7 days.</p>
			<p>If you don't want to accept this invitation, you can safely ignore this email.</p>
		`, inviterName, childName, acceptURL, acceptURL),
	}

	_, err := e.client.Emails.Send(params)
	return err
}