package service

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"log/slog"

	"github.com/go-mail/mail/v2"

	"github.com/vterdunov/learn-bank-app/internal/config"
	"github.com/vterdunov/learn-bank-app/internal/domain"
)

// EmailServiceImpl реализация EmailService
type EmailServiceImpl struct {
	dialer    *mail.Dialer
	from      string
	logger    *slog.Logger
	templates map[string]*template.Template
}

// NewEmailService создает новый экземпляр EmailService
func NewEmailService(cfg *config.Config, logger *slog.Logger) EmailService {
	d := mail.NewDialer(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password)
	d.TLSConfig = &tls.Config{
		ServerName:         cfg.SMTP.Host,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
	}

	// Загружаем шаблоны
	templates := make(map[string]*template.Template)
	templateFiles := map[string]string{
		"payment": "templates/email/payment_notification.tmpl",
		"credit":  "templates/email/credit_notification.tmpl",
		"overdue": "templates/email/overdue_notification.tmpl",
	}

	for name, file := range templateFiles {
		tmpl, err := template.ParseFiles(file)
		if err != nil {
			logger.Error("Failed to parse template", "name", name, "file", file, "error", err)
			continue
		}
		templates[name] = tmpl
	}

	return &EmailServiceImpl{
		dialer:    d,
		from:      cfg.SMTP.Username,
		logger:    logger,
		templates: templates,
	}
}

// SendPaymentNotification отправляет уведомление об успешном платеже
func (s *EmailServiceImpl) SendPaymentNotification(userEmail string, amount float64) error {
	subject := "Платеж успешно проведен"
	data := struct {
		Amount float64
	}{
		Amount: amount,
	}

	body, err := s.renderTemplate("payment", data)
	if err != nil {
		return fmt.Errorf("failed to render payment template: %w", err)
	}

	return s.sendEmail(userEmail, subject, body)
}

// SendCreditNotification отправляет уведомление о выдаче кредита
func (s *EmailServiceImpl) SendCreditNotification(userEmail string, credit *domain.Credit) error {
	subject := "Кредит успешно оформлен"
	body, err := s.renderTemplate("credit", credit)
	if err != nil {
		return fmt.Errorf("failed to render credit template: %w", err)
	}

	return s.sendEmail(userEmail, subject, body)
}

// SendOverdueNotification отправляет уведомление о просроченном платеже
func (s *EmailServiceImpl) SendOverdueNotification(userEmail string, payment *domain.PaymentSchedule) error {
	subject := "Просроченный платеж по кредиту"
	data := struct {
		PaymentAmount float64
		DueDate       string
		Status        string
	}{
		PaymentAmount: payment.PaymentAmount,
		DueDate:       payment.DueDate.Format("02.01.2006"),
		Status:        payment.Status,
	}

	body, err := s.renderTemplate("overdue", data)
	if err != nil {
		return fmt.Errorf("failed to render overdue template: %w", err)
	}

	return s.sendEmail(userEmail, subject, body)
}

// renderTemplate рендерит шаблон с данными
func (s *EmailServiceImpl) renderTemplate(templateName string, data interface{}) (string, error) {
	tmpl, ok := s.templates[templateName]
	if !ok {
		return "", fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// sendEmail отправляет email
func (s *EmailServiceImpl) sendEmail(to, subject, body string) error {
	m := mail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	if err := s.dialer.DialAndSend(m); err != nil {
		s.logger.Error("Failed to send email",
			"to", to,
			"subject", subject,
			"error", err)
		return fmt.Errorf("email sending failed: %w", err)
	}

	s.logger.Info("Email sent successfully",
		"to", to,
		"subject", subject)

	return nil
}
