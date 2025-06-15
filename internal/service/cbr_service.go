package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/etree"

	"github.com/vterdunov/learn-bank-app/internal/config"
)

// CBRServiceImpl реализация CBRService
type CBRServiceImpl struct {
	client     *http.Client
	logger     *slog.Logger
	serviceURL string
	bankMargin float64
}

// NewCBRService создает новый экземпляр CBRService
func NewCBRService(cfg *config.Config, logger *slog.Logger) CBRService {
	return &CBRServiceImpl{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:     logger,
		serviceURL: cfg.CBR.ServiceURL,
		bankMargin: cfg.CBR.BankMargin,
	}
}

// GetKeyRate получает ключевую ставку ЦБ РФ с добавлением маржи банка
func (s *CBRServiceImpl) GetKeyRate(ctx context.Context) (float64, error) {
	s.logger.Info("Requesting key rate from CBR")

	// Формируем SOAP запрос
	soapRequest := s.buildSOAPRequest()

	// Отправляем запрос
	rawBody, err := s.sendRequest(ctx, soapRequest)
	if err != nil {
		s.logger.Error("Failed to send request to CBR", "error", err)
		return 0, fmt.Errorf("failed to send CBR request: %w", err)
	}

	// Парсим XML ответ
	rate, err := s.parseXMLResponse(rawBody)
	if err != nil {
		s.logger.Error("Failed to parse CBR response", "error", err)
		return 0, fmt.Errorf("failed to parse CBR response: %w", err)
	}

	// Добавляем маржу банка
	finalRate := rate + s.bankMargin

	s.logger.Info("Successfully got key rate from CBR",
		"cbr_rate", rate,
		"bank_margin", s.bankMargin,
		"final_rate", finalRate)

	return finalRate, nil
}

// buildSOAPRequest формирует SOAP запрос для получения ключевой ставки
func (s *CBRServiceImpl) buildSOAPRequest() string {
	fromDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	toDate := time.Now().Format("2006-01-02")

	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
    <soap12:Body>
        <KeyRate xmlns="http://web.cbr.ru/">
            <fromDate>%s</fromDate>
            <ToDate>%s</ToDate>
        </KeyRate>
    </soap12:Body>
</soap12:Envelope>`, fromDate, toDate)
}

// sendRequest отправляет SOAP запрос к ЦБ РФ
func (s *CBRServiceImpl) sendRequest(ctx context.Context, soapRequest string) ([]byte, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		s.serviceURL,
		bytes.NewBuffer([]byte(soapRequest)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем заголовки для SOAP
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://web.cbr.ru/KeyRate")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CBR service returned status: %d", resp.StatusCode)
	}

	// Читаем ответ
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return rawBody, nil
}

// parseXMLResponse парсит XML ответ от ЦБ РФ и извлекает ключевую ставку
func (s *CBRServiceImpl) parseXMLResponse(rawBody []byte) (float64, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(rawBody); err != nil {
		return 0, fmt.Errorf("failed to parse XML: %w", err)
	}

	// Ищем элементы ключевой ставки в XML
	// Структура: //diffgram/KeyRate/KR
	krElements := doc.FindElements("//diffgram/KeyRate/KR")
	if len(krElements) == 0 {
		return 0, errors.New("key rate data not found in response")
	}

	// Берем последний (самый актуальный) элемент
	latestKR := krElements[len(krElements)-1]

	// Ищем элемент Rate
	rateElement := latestKR.FindElement("./Rate")
	if rateElement == nil {
		return 0, errors.New("rate element not found")
	}

	rateStr := strings.TrimSpace(rateElement.Text())
	if rateStr == "" {
		return 0, errors.New("rate value is empty")
	}

	// Конвертируем строку в число
	rate, err := strconv.ParseFloat(rateStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse rate value '%s': %w", rateStr, err)
	}

	if rate < 0 {
		return 0, fmt.Errorf("invalid rate value: %f", rate)
	}

	return rate, nil
}
