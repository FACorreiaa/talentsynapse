package auth

import (
	"fmt"
	"os"

	"github.com/resend/resend-go/v3"
)

// EmailService handles sending emails via Resend
type EmailService struct {
	client *resend.Client
	from   string
	appURL string
}

// NewEmailService creates a new email service
func NewEmailService() *EmailService {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		// For development, log warning but don't fail
		fmt.Println("WARNING: RESEND_API_KEY not set, emails will not be sent")
	}

	from := os.Getenv("EMAIL_FROM")
	if from == "" {
		from = "TalentSynapse <noreply@TalentSynapse.com>"
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:8080"
	}

	client := resend.NewClient(apiKey)

	return &EmailService{
		client: client,
		from:   from,
		appURL: appURL,
	}
}

// SendVerificationEmail sends an email verification link to a user
func (s *EmailService) SendVerificationEmail(to, username, token string) error {
	if s.client == nil {
		fmt.Printf("Email simulation: Verification email would be sent to %s\n", to)
		fmt.Printf("Verification URL: %s/verify-email?token=%s\n", s.appURL, token)
		return nil
	}

	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.appURL, token)

	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: "Verify your TalentSynapse account",
		Html:    s.generateVerificationEmailHTML(username, verifyURL),
	}

	_, err := s.client.Emails.Send(params)
	return err
}

// SendMFAEmail sends an MFA setup confirmation email
func (s *EmailService) SendMFAEmail(to, username string, enabled bool) error {
	if s.client == nil {
		action := "enabled"
		if !enabled {
			action = "disabled"
		}
		fmt.Printf("Email simulation: MFA %s email would be sent to %s\n", action, to)
		return nil
	}

	subject := "Two-Factor Authentication Enabled"
	html := s.generateMFAEnabledHTML(username)

	if !enabled {
		subject = "Two-Factor Authentication Disabled"
		html = s.generateMFADisabledHTML(username)
	}

	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: subject,
		Html:    html,
	}

	_, err := s.client.Emails.Send(params)
	return err
}

// SendPasswordResetEmail sends a password reset link to a user
func (s *EmailService) SendPasswordResetEmail(to, username, token string) error {
	if s.client == nil {
		fmt.Printf("Email simulation: Password reset email would be sent to %s\n", to)
		fmt.Printf("Reset URL: %s/reset-password?token=%s\n", s.appURL, token)
		return nil
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.appURL, token)

	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: "Reset your TalentSynapse password",
		Html:    s.generatePasswordResetHTML(username, resetURL),
	}

	_, err := s.client.Emails.Send(params)
	return err
}

// generateVerificationEmailHTML generates the HTML for email verification
func (s *EmailService) generateVerificationEmailHTML(username, verifyURL string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Verify Your Email</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #f5f5f5;">
    <table role="presentation" style="width: 100%%; border-collapse: collapse;">
        <tr>
            <td align="center" style="padding: 40px 0;">
                <table role="presentation" style="width: 600px; max-width: 600px; background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 40px; text-align: center; border-radius: 8px 8px 0 0;">
                            <h1 style="margin: 0; color: #ffffff; font-size: 28px; font-weight: bold;">
                                TalentSynapse
                            </h1>
                        </td>
                    </tr>

                    <!-- Body -->
                    <tr>
                        <td style="padding: 40px;">
                            <h2 style="margin: 0 0 20px 0; color: #1a202c; font-size: 24px; font-weight: 600;">
                                Welcome to TalentSynapse, %s! 🎉
                            </h2>

                            <p style="margin: 0 0 20px 0; color: #4a5568; font-size: 16px; line-height: 1.6;">
                                Thank you for joining our skill exchange community. To get started, please verify your email address by clicking the button below:
                            </p>

                            <!-- Button -->
                            <table role="presentation" style="margin: 30px 0;">
                                <tr>
                                    <td style="text-align: center;">
                                        <a href="%s" style="display: inline-block; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: #ffffff; text-decoration: none; padding: 14px 32px; border-radius: 6px; font-weight: 600; font-size: 16px;">
                                            Verify Email Address
                                        </a>
                                    </td>
                                </tr>
                            </table>

                            <p style="margin: 20px 0; color: #718096; font-size: 14px; line-height: 1.6;">
                                Or copy and paste this link into your browser:
                            </p>
                            <p style="margin: 0 0 20px 0; color: #667eea; font-size: 14px; word-break: break-all;">
                                %s
                            </p>

                            <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #e2e8f0;">
                                <p style="margin: 0; color: #718096; font-size: 13px; line-height: 1.6;">
                                    ⏰ This link will expire in 24 hours for security purposes.
                                </p>
                                <p style="margin: 10px 0 0 0; color: #718096; font-size: 13px; line-height: 1.6;">
                                    If you didn't create an account with TalentSynapse, you can safely ignore this email.
                                </p>
                            </div>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #f7fafc; padding: 30px; text-align: center; border-radius: 0 0 8px 8px;">
                            <p style="margin: 0 0 10px 0; color: #4a5568; font-size: 16px; font-weight: 600;">
                                Happy Learning! 📚
                            </p>
                            <p style="margin: 0; color: #718096; font-size: 13px;">
                                The TalentSynapse Team
                            </p>
                        </td>
                    </tr>
                </table>

                <!-- Footer Text -->
                <p style="margin: 20px 0 0 0; color: #a0aec0; font-size: 12px; text-align: center;">
                    © 2026 TalentSynapse. All rights reserved.
                </p>
            </td>
        </tr>
    </table>
</body>
</html>
`, username, verifyURL, verifyURL)
}

// generateMFAEnabledHTML generates the HTML for MFA enabled notification
func (s *EmailService) generateMFAEnabledHTML(username string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MFA Enabled</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #f5f5f5;">
    <table role="presentation" style="width: 100%%; border-collapse: collapse;">
        <tr>
            <td align="center" style="padding: 40px 0;">
                <table role="presentation" style="width: 600px; max-width: 600px; background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #10b981 0%%, #059669 100%%); padding: 40px; text-align: center; border-radius: 8px 8px 0 0;">
                            <div style="font-size: 48px; margin-bottom: 10px;">🔐</div>
                            <h1 style="margin: 0; color: #ffffff; font-size: 28px; font-weight: bold;">
                                MFA Enabled
                            </h1>
                        </td>
                    </tr>

                    <!-- Body -->
                    <tr>
                        <td style="padding: 40px;">
                            <h2 style="margin: 0 0 20px 0; color: #1a202c; font-size: 24px; font-weight: 600;">
                                Hi %s,
                            </h2>

                            <p style="margin: 0 0 20px 0; color: #4a5568; font-size: 16px; line-height: 1.6;">
                                Two-factor authentication (MFA) has been successfully enabled on your TalentSynapse account.
                            </p>

                            <div style="background-color: #d1fae5; border-left: 4px solid #10b981; padding: 16px; margin: 20px 0; border-radius: 4px;">
                                <p style="margin: 0; color: #065f46; font-size: 14px; line-height: 1.6;">
                                    <strong>✓ Your account is now more secure!</strong><br>
                                    You'll need your authenticator app to sign in from now on.
                                </p>
                            </div>

                            <p style="margin: 20px 0; color: #4a5568; font-size: 16px; line-height: 1.6;">
                                <strong>What this means:</strong>
                            </p>
                            <ul style="margin: 0 0 20px 0; color: #4a5568; font-size: 16px; line-height: 1.8; padding-left: 20px;">
                                <li>You'll need a code from your authenticator app when signing in</li>
                                <li>Your backup codes can be used if you lose access to your authenticator</li>
                                <li>You can disable MFA anytime from your security settings</li>
                            </ul>

                            <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #e2e8f0;">
                                <p style="margin: 0; color: #e53e3e; font-size: 14px; line-height: 1.6;">
                                    ⚠️ <strong>Didn't enable MFA?</strong> If you didn't make this change, please secure your account immediately by resetting your password.
                                </p>
                            </div>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #f7fafc; padding: 30px; text-align: center; border-radius: 0 0 8px 8px;">
                            <p style="margin: 0; color: #718096; font-size: 13px;">
                                The TalentSynapse Security Team
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, username)
}

// generateMFADisabledHTML generates the HTML for MFA disabled notification
func (s *EmailService) generateMFADisabledHTML(username string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MFA Disabled</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #f5f5f5;">
    <table role="presentation" style="width: 100%%; border-collapse: collapse;">
        <tr>
            <td align="center" style="padding: 40px 0;">
                <table role="presentation" style="width: 600px; max-width: 600px; background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #f59e0b 0%%, #d97706 100%%); padding: 40px; text-align: center; border-radius: 8px 8px 0 0;">
                            <div style="font-size: 48px; margin-bottom: 10px;">🔓</div>
                            <h1 style="margin: 0; color: #ffffff; font-size: 28px; font-weight: bold;">
                                MFA Disabled
                            </h1>
                        </td>
                    </tr>

                    <!-- Body -->
                    <tr>
                        <td style="padding: 40px;">
                            <h2 style="margin: 0 0 20px 0; color: #1a202c; font-size: 24px; font-weight: 600;">
                                Hi %s,
                            </h2>

                            <p style="margin: 0 0 20px 0; color: #4a5568; font-size: 16px; line-height: 1.6;">
                                Two-factor authentication (MFA) has been disabled on your TalentSynapse account.
                            </p>

                            <div style="background-color: #fef3c7; border-left: 4px solid #f59e0b; padding: 16px; margin: 20px 0; border-radius: 4px;">
                                <p style="margin: 0; color: #92400e; font-size: 14px; line-height: 1.6;">
                                    <strong>⚠️ Security Notice</strong><br>
                                    Your account is now less secure without MFA. We recommend re-enabling it.
                                </p>
                            </div>

                            <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #e2e8f0;">
                                <p style="margin: 0; color: #e53e3e; font-size: 14px; line-height: 1.6;">
                                    ⚠️ <strong>Didn't disable MFA?</strong> If you didn't make this change, please secure your account immediately by resetting your password and re-enabling MFA.
                                </p>
                            </div>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #f7fafc; padding: 30px; text-align: center; border-radius: 0 0 8px 8px;">
                            <p style="margin: 0; color: #718096; font-size: 13px;">
                                The TalentSynapse Security Team
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, username)
}

// generatePasswordResetHTML generates the HTML for password reset
func (s *EmailService) generatePasswordResetHTML(username, resetURL string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset Your Password</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #f5f5f5;">
    <table role="presentation" style="width: 100%%; border-collapse: collapse;">
        <tr>
            <td align="center" style="padding: 40px 0;">
                <table role="presentation" style="width: 600px; max-width: 600px; background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 40px; text-align: center; border-radius: 8px 8px 0 0;">
                            <h1 style="margin: 0; color: #ffffff; font-size: 28px; font-weight: bold;">
                                TalentSynapse
                            </h1>
                        </td>
                    </tr>

                    <!-- Body -->
                    <tr>
                        <td style="padding: 40px;">
                            <h2 style="margin: 0 0 20px 0; color: #1a202c; font-size: 24px; font-weight: 600;">
                                Reset Your Password
                            </h2>

                            <p style="margin: 0 0 20px 0; color: #4a5568; font-size: 16px; line-height: 1.6;">
                                Hi %s,
                            </p>

                            <p style="margin: 0 0 20px 0; color: #4a5568; font-size: 16px; line-height: 1.6;">
                                We received a request to reset your password. Click the button below to create a new password:
                            </p>

                            <!-- Button -->
                            <table role="presentation" style="margin: 30px 0;">
                                <tr>
                                    <td style="text-align: center;">
                                        <a href="%s" style="display: inline-block; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: #ffffff; text-decoration: none; padding: 14px 32px; border-radius: 6px; font-weight: 600; font-size: 16px;">
                                            Reset Password
                                        </a>
                                    </td>
                                </tr>
                            </table>

                            <p style="margin: 20px 0; color: #718096; font-size: 14px; line-height: 1.6;">
                                Or copy and paste this link into your browser:
                            </p>
                            <p style="margin: 0 0 20px 0; color: #667eea; font-size: 14px; word-break: break-all;">
                                %s
                            </p>

                            <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #e2e8f0;">
                                <p style="margin: 0; color: #718096; font-size: 13px; line-height: 1.6;">
                                    ⏰ This link will expire in 1 hour for security purposes.
                                </p>
                                <p style="margin: 10px 0 0 0; color: #718096; font-size: 13px; line-height: 1.6;">
                                    If you didn't request a password reset, you can safely ignore this email. Your password will not be changed.
                                </p>
                            </div>
                        </td>
                    </tr>

                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #f7fafc; padding: 30px; text-align: center; border-radius: 0 0 8px 8px;">
                            <p style="margin: 0; color: #718096; font-size: 13px;">
                                The TalentSynapse Team
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`, username, resetURL, resetURL)
}
