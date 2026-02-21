package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type appointmentIntegrationSuite struct {
	suite.Suite

	client    *http.Client
	apiURL    string
	authToken string

	patientID string
	doctorID  string
	unitID    string
	serviceID string

	treatmentType string
}

func TestAppointmentIntegrationSuite(t *testing.T) {
	suite.Run(t, new(appointmentIntegrationSuite))
}

func (s *appointmentIntegrationSuite) SetupSuite() {
	requiredVars := []string{
		"INTEGRATION_API_URL",
		"INTEGRATION_AUTH_TOKEN",
		"INTEGRATION_PATIENT_ID",
		"INTEGRATION_DOCTOR_ID",
		"INTEGRATION_UNIT_ID",
		"INTEGRATION_SERVICE_ID",
	}

	missing := missingEnvVars(requiredVars)
	if len(missing) > 0 {
		s.T().Fatalf("integration environment not configured, missing required vars: %s", strings.Join(missing, ", "))
	}

	s.apiURL = strings.TrimRight(os.Getenv("INTEGRATION_API_URL"), "/")
	s.authToken = os.Getenv("INTEGRATION_AUTH_TOKEN")
	s.patientID = os.Getenv("INTEGRATION_PATIENT_ID")
	s.doctorID = os.Getenv("INTEGRATION_DOCTOR_ID")
	s.unitID = os.Getenv("INTEGRATION_UNIT_ID")
	s.serviceID = os.Getenv("INTEGRATION_SERVICE_ID")

	s.treatmentType = os.Getenv("INTEGRATION_TREATMENT_TYPE")
	if s.treatmentType == "" {
		s.treatmentType = "limpieza"
	}

	s.client = &http.Client{Timeout: 30 * time.Second}
}

func (s *appointmentIntegrationSuite) postJSON(path string, payload any) (*http.Response, envelopeResponse, error) {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, envelopeResponse{}, err
	}

	url := s.apiURL + path
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
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

	var parsed envelopeResponse
	decodeErr := json.NewDecoder(resp.Body).Decode(&parsed)
	if decodeErr != nil {
		return resp, envelopeResponse{}, decodeErr
	}

	return resp, parsed, nil
}

func missingEnvVars(keys []string) []string {
	missing := make([]string, 0)
	for _, key := range keys {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			missing = append(missing, key)
		}
	}

	return missing
}

type envelopeResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *errorResponse  `json:"error,omitempty"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type createAppointmentResponse struct {
	ID          string  `json:"id"`
	PatientID   *string `json:"patient_id,omitempty"`
	DoctorID    *string `json:"doctor_id,omitempty"`
	UnitID      *string `json:"unit_id,omitempty"`
	ServiceID   *string `json:"service_id,omitempty"`
	Status      string  `json:"status"`
	StartTime   string  `json:"start_time"`
	EndTime     string  `json:"end_time"`
	PatientName string  `json:"patient_name"`
}

func requireNoAPIError(t *testing.T, envelope envelopeResponse) {
	require.True(t, envelope.Success, "expected success=true, got error: %+v", envelope.Error)
	require.Nil(t, envelope.Error, "expected no error payload")
	require.NotEmpty(t, envelope.Data, "expected response data")
}
