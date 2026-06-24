package server

import (
	"fmt"
	"strings"

	"github.com/DatanoiseTV/tinyice/config"
	"github.com/DatanoiseTV/tinyice/logger"
	mail "github.com/wneessen/go-mail"
)

func (s *Server) adminNotificationEmails() []string {
	seen := map[string]bool{}
	var out []string

	if e := strings.TrimSpace(s.Config.AdminEmail); e != "" {
		seen[e] = true
		out = append(out, e)
	}
	for _, user := range s.Config.Users {
		if user == nil || user.Role != config.RoleSuperAdmin {
			continue
		}
		for _, e := range user.LinkedEmails {
			e = strings.TrimSpace(e)
			if e != "" && !seen[e] {
				seen[e] = true
				out = append(out, e)
			}
		}
	}
	return out
}

func (s *Server) notifyAdminsEvent(event, subject, body string) {
	if s.Config.SMTP == nil || !s.Config.SMTP.WantsNotify(event) {
		return
	}
	recipients := s.adminNotificationEmails()
	if len(recipients) == 0 {
		return
	}
	for _, to := range recipients {
		if err := s.sendEmail(to, subject, body); err != nil {
			logger.L.Warnw("Failed to send notification email", "event", event, "to", to, "error", err)
		}
	}
}

func (s *Server) notifyAdminsNewPendingUser(pending *config.PendingUser) {
	subject := fmt.Sprintf("[TinyIce] New access request from %s", pending.Email)
	body := fmt.Sprintf("A new user has requested access to your TinyIce server.\n\n"+
		"Name: %s\n"+
		"Email: %s\n"+
		"Provider: %s\n"+
		"Requested: %s\n\n"+
		"Log in to your admin panel to approve or deny this request.",
		pending.Name, pending.Email, pending.Provider, pending.RequestedAt)
	s.notifyAdminsEvent("pending_user", subject, body)
}

func (s *Server) notifyAdminsSourceConnect(mount, ip, ua, name string) {
	subject := fmt.Sprintf("[TinyIce] Source connected on %s", mount)
	body := fmt.Sprintf("A source has connected to your TinyIce server.\n\n"+
		"Mount: %s\n"+
		"IP: %s\n"+
		"User-Agent: %s\n"+
		"Name: %s\n",
		mount, ip, ua, name)
	s.notifyAdminsEvent("source_connect", subject, body)
}

func (s *Server) notifyAdminsSourceDisconnect(mount string) {
	subject := fmt.Sprintf("[TinyIce] Source disconnected from %s", mount)
	body := fmt.Sprintf("The source on mount %s has disconnected.", mount)
	s.notifyAdminsEvent("source_disconnect", subject, body)
}

func (s *Server) notifyAdminsSecurityLockout(ip, reason, until, details string) {
	subject := fmt.Sprintf("[TinyIce] Security lockout: %s", ip)
	body := fmt.Sprintf("An IP address was locked out on your TinyIce server.\n\n"+
		"IP: %s\n"+
		"Reason: %s\n"+
		"Until: %s\n"+
		"Details: %s\n",
		ip, reason, until, details)
	s.notifyAdminsEvent("security_lockout", subject, body)
}

func (s *Server) sendEmail(to, subject, body string) error {
	smtp := s.Config.SMTP
	if smtp == nil || !smtp.Enabled {
		return fmt.Errorf("SMTP not configured")
	}

	m := mail.NewMsg()
	if err := m.From(smtp.From); err != nil {
		return err
	}
	if err := m.To(to); err != nil {
		return err
	}
	m.Subject(subject)
	m.SetBodyString(mail.TypeTextPlain, body)

	port := smtp.Port
	if port == 0 {
		port = 587
	}

	c, err := mail.NewClient(smtp.Host,
		mail.WithPort(port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(smtp.Username),
		mail.WithPassword(smtp.Password),
		mail.WithTLSPortPolicy(mail.TLSMandatory),
	)
	if err != nil {
		return err
	}

	return c.DialAndSend(m)
}
