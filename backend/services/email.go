package services

import (
	"fmt"
	"os"

	"github.com/booktracker/backend/models"
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

func (e *EmailService) SendPasswordResetEmail(email, firstName, resetToken string) error {
	if e.client == nil {
		// Development mode - just log
		fmt.Printf("📧 [DEV] Password reset email for %s:\n", email)
		fmt.Printf("   Token: %s\n", resetToken)
		fmt.Printf("   URL: http://localhost:5173/reset-password?token=%s\n", resetToken)
		return nil
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173" // fallback for development
	}
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, resetToken)
	
	params := &resend.SendEmailRequest{
		From:    "Book Tracker <noreply@booktracker.rustyphillips.net>",
		To:      []string{email},
		Subject: "Reset your password",
		Html: fmt.Sprintf(`
			<h1>Password Reset Request</h1>
			<p>Hi %s,</p>
			<p>We received a request to reset your password for your Book Tracker account.</p>
			<p>Click the link below to reset your password:</p>
			<p><a href="%s">Reset Password</a></p>
			<p>If the button doesn't work, copy and paste this URL into your browser:</p>
			<p>%s</p>
			<p>This link will expire in 1 hour.</p>
			<p>If you didn't request this password reset, you can safely ignore this email.</p>
		`, firstName, resetURL, resetURL),
	}

	_, err := e.client.Emails.Send(params)
	return err
}

func (e *EmailService) SendSystemInvitationEmail(email, inviterName, token string) error {
	if e.client == nil {
		// Development mode - just log
		fmt.Printf("📧 [DEV] System invitation email for %s:\n", email)
		fmt.Printf("   Inviter: %s\n", inviterName)
		fmt.Printf("   Token: %s\n", token)
		fmt.Printf("   URL: http://localhost:5173/accept-invitation?token=%s\n", token)
		return nil
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173" // fallback for development
	}
	acceptURL := fmt.Sprintf("%s/accept-invitation?token=%s", frontendURL, token)
	
	params := &resend.SendEmailRequest{
		From:    "Book Tracker <noreply@booktracker.rustyphillips.net>",
		To:      []string{email},
		Subject: fmt.Sprintf("%s has invited you to join Book Tracker", inviterName),
		Html: fmt.Sprintf(`
			<h1>You've been invited to Book Tracker!</h1>
			<p>%s has invited you to join Book Tracker to help track reading progress.</p>
			<p>Book Tracker is a simple way to log and monitor children's reading activities, celebrate achievements, and encourage a love of reading.</p>
			<p>Click the link below to create your account and get started:</p>
			<p><a href="%s" style="background-color: #4F46E5; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Join Book Tracker</a></p>
			<p>If the button doesn't work, copy and paste this URL into your browser:</p>
			<p>%s</p>
			<p>Once you create your account, you'll automatically have access to the children %s has shared with you.</p>
			<p>If you don't want to join, you can safely ignore this email.</p>
		`, inviterName, acceptURL, acceptURL, inviterName),
	}

	_, err := e.client.Emails.Send(params)
	return err
}

var emailService *EmailService

func init() {
	emailService = NewEmailService()
}

// SendInvitationEmail sends an invitation email using the models
func SendInvitationEmail(email, token string, inviter *models.User, child *models.Child) error {
	inviterName := fmt.Sprintf("%s %s", inviter.FirstName, inviter.LastName)
	childName := fmt.Sprintf("%s %s", child.FirstName, child.LastName)
	return emailService.SendInvitationEmail(email, inviterName, childName, token)
}

// SendSystemInvitationEmail sends a general system invitation (not child-specific)
func SendSystemInvitationEmail(email, token string, inviter *models.User) error {
	inviterName := fmt.Sprintf("%s %s", inviter.FirstName, inviter.LastName)
	return emailService.SendSystemInvitationEmail(email, inviterName, token)
}