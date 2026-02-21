package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/stretchr/testify/require"
)

type cashSessionResponse struct {
	ID                 string `json:"id"`
	StartingFloatCents int64  `json:"starting_float_cents"`
	Status             string `json:"status"`
}

type cashSessionListItem struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type accountBalanceResponse struct {
	BalanceDueCents int64 `json:"balance_due_cents"`
}

type accountEntryResponse struct {
	ID string `json:"id"`
}

type sessionDetailsResponse struct {
	Session struct {
		ID                 string `json:"id"`
		StartingFloatCents int64  `json:"starting_float_cents"`
		Status             string `json:"status"`
	} `json:"session"`
	ExpectedAmounts map[string]int64            `json:"expected_amounts"`
	PaymentSummary  map[string]map[string]int64 `json:"payment_summary"`
}

func (s *appointmentIntegrationSuite) TestAccountingIntegrationFlow() {
	sessionID := s.openCashSessionWithFloat(50000)

	baseStart := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Minute)

	appointment1ID := s.createAppointmentAt(baseStart, baseStart.Add(30*time.Minute))

	s.createServiceCharge(appointment1ID, map[string]any{
		"service_id":     s.serviceID,
		"doctor_id":      s.doctorID,
		"doctor_type":    "internal",
		"amount_cents":   int64(150000),
		"commission_pct": 30.0,
		"currency":       "MXN",
		"description":    "service 1 internal charge",
	})
	s.assertAppointmentBalance(appointment1ID, 150000)

	charge2ID := s.createServiceCharge(appointment1ID, map[string]any{
		"service_id":                s.serviceID,
		"doctor_id":                 s.doctorID,
		"doctor_type":               "external",
		"amount_cents":              int64(150000),
		"external_doctor_fee_cents": int64(50000),
		"currency":                  "MXN",
		"description":               "service 2 external charge",
	})
	s.assertAppointmentBalance(appointment1ID, 300000)

	s.createCorrection(appointment1ID, charge2ID, "full rollback for charge 2")
	s.assertAppointmentBalance(appointment1ID, 150000)

	s.createServiceCharge(appointment1ID, map[string]any{
		"service_id":                s.serviceID,
		"doctor_id":                 s.doctorID,
		"doctor_type":               "external",
		"amount_cents":              int64(100000),
		"external_doctor_fee_cents": int64(50000),
		"currency":                  "MXN",
		"description":               "service 2 external charge adjusted",
	})
	s.assertAppointmentBalance(appointment1ID, 250000)

	s.createPayment(appointment1ID, map[string]any{
		"amount_cents":   int64(50000),
		"currency":       "MXN",
		"payment_method": "cash",
		"description":    "payment 1 cash mxn",
	})

	s.createPayment(appointment1ID, map[string]any{
		"amount_cents":   int64(2500),
		"currency":       "USD",
		"payment_method": "cash",
		"exchange_rate":  20.0,
		"description":    "payment 2 cash usd",
	})
	s.assertAppointmentBalance(appointment1ID, 150000)

	s.createPayment(appointment1ID, map[string]any{
		"amount_cents":   int64(100000),
		"currency":       "MXN",
		"payment_method": "card",
		"description":    "payment 3 card mxn",
	})
	s.assertAppointmentBalance(appointment1ID, 50000)

	appointment2ID := s.createAppointmentAt(baseStart.Add(2*time.Hour), baseStart.Add(150*time.Minute))

	s.createServiceCharge(appointment2ID, map[string]any{
		"service_id":     s.serviceID,
		"doctor_id":      s.doctorID,
		"doctor_type":    "internal",
		"amount_cents":   int64(50000),
		"commission_pct": 30.0,
		"currency":       "MXN",
		"description":    "appointment 2 internal charge",
	})

	s.createPayment(appointment2ID, map[string]any{
		"amount_cents":   int64(50000),
		"currency":       "MXN",
		"payment_method": "card",
		"description":    "appointment 2 card payment",
	})
	s.assertAppointmentBalance(appointment2ID, 0)

	details := s.getCashSessionDetails(sessionID)
	require.Equal(s.T(), int64(50000), details.Session.StartingFloatCents)
	require.Equal(s.T(), "open", details.Session.Status)

	require.Equal(s.T(), int64(150000), mapValueOrZero(details.ExpectedAmounts, "MXN"))
	require.Equal(s.T(), int64(2500), mapValueOrZero(details.ExpectedAmounts, "USD"))

	require.Equal(s.T(), int64(150000), nestedMapValueOrZero(details.PaymentSummary, "card", "MXN"))
	require.Equal(s.T(), int64(0), nestedMapValueOrZero(details.PaymentSummary, "card", "USD"))
}

func (s *appointmentIntegrationSuite) doJSON(method, path string, payload any) (*http.Response, envelopeResponse, error) {
	var body *bytes.Buffer
	if payload != nil {
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, envelopeResponse{}, err
		}
		body = bytes.NewBuffer(bodyBytes)
	} else {
		body = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, s.apiURL+path, body)
	if err != nil {
		return nil, envelopeResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.authToken))

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, envelopeResponse{}, err
	}
	defer resp.Body.Close()

	var envelope envelopeResponse
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return resp, envelopeResponse{}, err
	}

	return resp, envelope, nil
}

func (s *appointmentIntegrationSuite) openCashSessionWithFloat(startingFloatCents int64) string {
	openNew := func() (*http.Response, envelopeResponse, error) {
		payload := map[string]any{
			"opening_type":         "manual",
			"starting_float_cents": startingFloatCents,
			"notes":                "integration flow",
		}
		return s.doJSON(http.MethodPost, "/api/v1/cash-sessions/open", payload)
	}

	resp, envelope, err := openNew()
	require.NoError(s.T(), err)

	if resp.StatusCode != http.StatusCreated {
		if envelope.Error == nil || !strings.Contains(envelope.Error.Message, "already has an open cash session") {
			require.Equal(s.T(), http.StatusCreated, resp.StatusCode, "open session failed: %+v", envelope.Error)
		}

		openSessions := s.listOpenCashSessions()
		for _, session := range openSessions {
			s.closeCashSession(session.ID)
		}

		resp, envelope, err = openNew()
		require.NoError(s.T(), err)
		require.Equal(s.T(), http.StatusCreated, resp.StatusCode, "open session retry failed: %+v", envelope.Error)
	}

	requireNoAPIError(s.T(), envelope)

	var session cashSessionResponse
	err = json.Unmarshal(envelope.Data, &session)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), session.ID)
	return session.ID
}

func (s *appointmentIntegrationSuite) listOpenCashSessions() []cashSessionListItem {
	resp, envelope, err := s.doJSON(http.MethodGet, "/api/v1/cash-sessions?page=1&limit=100", nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode, "list sessions failed: %+v", envelope.Error)
	requireNoAPIError(s.T(), envelope)

	var sessions []cashSessionListItem
	err = json.Unmarshal(envelope.Data, &sessions)
	require.NoError(s.T(), err)

	openSessions := make([]cashSessionListItem, 0)
	for _, session := range sessions {
		if session.Status == "open" {
			openSessions = append(openSessions, session)
		}
	}

	return openSessions
}

func (s *appointmentIntegrationSuite) closeCashSession(sessionID string) {
	resp, envelope, err := s.doJSON(http.MethodPost, fmt.Sprintf("/api/v1/cash-sessions/%s/close", sessionID), map[string]any{})
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode, "close session failed: %+v", envelope.Error)
	require.True(s.T(), envelope.Success)
}

func (s *appointmentIntegrationSuite) createAppointmentAt(start, end time.Time) string {
	req := createAppointmentRequest{
		PatientID:     s.patientID,
		DoctorID:      s.doctorID,
		UnitID:        s.unitID,
		ServiceID:     s.serviceID,
		StartTime:     start.Format(time.RFC3339Nano),
		EndTime:       end.Format(time.RFC3339Nano),
		TreatmentType: s.treatmentType,
	}

	resp, envelope, err := s.postJSON("/api/v1/appointments", req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode, "create appointment failed: %+v", envelope.Error)
	requireNoAPIError(s.T(), envelope)

	var created createAppointmentResponse
	err = json.Unmarshal(envelope.Data, &created)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), created.ID)

	return created.ID
}

func (s *appointmentIntegrationSuite) createServiceCharge(appointmentID string, payload map[string]any) string {
	resp, envelope, err := s.doJSON(http.MethodPost, fmt.Sprintf("/api/v1/appointments/%s/account/charges", appointmentID), payload)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode, "create service charge failed: %+v", envelope.Error)
	requireNoAPIError(s.T(), envelope)

	var entry accountEntryResponse
	err = json.Unmarshal(envelope.Data, &entry)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), entry.ID)

	return entry.ID
}

func (s *appointmentIntegrationSuite) createCorrection(appointmentID, entryID, description string) string {
	payload := map[string]any{"description": description}
	resp, envelope, err := s.doJSON(http.MethodPost, fmt.Sprintf("/api/v1/appointments/%s/account/entries/%s/correct", appointmentID, entryID), payload)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode, "create correction failed: %+v", envelope.Error)
	requireNoAPIError(s.T(), envelope)

	var entry accountEntryResponse
	err = json.Unmarshal(envelope.Data, &entry)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), entry.ID)

	return entry.ID
}

func (s *appointmentIntegrationSuite) createPayment(appointmentID string, payload map[string]any) string {
	resp, envelope, err := s.doJSON(http.MethodPost, fmt.Sprintf("/api/v1/appointments/%s/account/payments", appointmentID), payload)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode, "create payment failed: %+v", envelope.Error)
	requireNoAPIError(s.T(), envelope)

	var entry accountEntryResponse
	err = json.Unmarshal(envelope.Data, &entry)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), entry.ID)

	return entry.ID
}

func (s *appointmentIntegrationSuite) assertAppointmentBalance(appointmentID string, expectedBalanceCents int64) {
	resp, envelope, err := s.doJSON(http.MethodGet, fmt.Sprintf("/api/v1/appointments/%s/account/balance", appointmentID), nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode, "get balance failed: %+v", envelope.Error)
	requireNoAPIError(s.T(), envelope)

	var balance accountBalanceResponse
	err = json.Unmarshal(envelope.Data, &balance)
	require.NoError(s.T(), err)
	require.Equal(s.T(), expectedBalanceCents, balance.BalanceDueCents)
}

func (s *appointmentIntegrationSuite) getCashSessionDetails(sessionID string) sessionDetailsResponse {
	resp, envelope, err := s.doJSON(http.MethodGet, fmt.Sprintf("/api/v1/cash-sessions/%s", sessionID), nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode, "get session details failed: %+v", envelope.Error)
	requireNoAPIError(s.T(), envelope)

	var details sessionDetailsResponse
	err = json.Unmarshal(envelope.Data, &details)
	require.NoError(s.T(), err)

	return details
}

func mapValueOrZero(values map[string]int64, key string) int64 {
	if values == nil {
		return 0
	}
	if value, ok := values[key]; ok {
		return value
	}
	return 0
}

func nestedMapValueOrZero(values map[string]map[string]int64, key1, key2 string) int64 {
	if values == nil {
		return 0
	}
	nested, ok := values[key1]
	if !ok {
		return 0
	}
	if value, ok := nested[key2]; ok {
		return value
	}
	return 0
}
