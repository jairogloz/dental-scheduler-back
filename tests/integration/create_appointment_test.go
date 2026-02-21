package integration_test

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/stretchr/testify/require"
)

type createAppointmentRequest struct {
	PatientID     string `json:"patient_id"`
	DoctorID      string `json:"doctor_id"`
	UnitID        string `json:"unit_id"`
	ServiceID     string `json:"service_id"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	TreatmentType string `json:"treatment_type"`
}

func (s *appointmentIntegrationSuite) TestCreateAppointment() {
	start := time.Now().UTC().Add(24 * time.Hour)
	end := start.Add(30 * time.Minute)

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
	require.Equal(s.T(), http.StatusCreated, resp.StatusCode)

	requireNoAPIError(s.T(), envelope)

	var created createAppointmentResponse
	err = json.Unmarshal(envelope.Data, &created)
	require.NoError(s.T(), err)

	require.NotEmpty(s.T(), created.ID)
	require.Equal(s.T(), "scheduled", created.Status)

	require.NotNil(s.T(), created.PatientID)
	require.NotNil(s.T(), created.DoctorID)
	require.NotNil(s.T(), created.UnitID)
	require.NotNil(s.T(), created.ServiceID)

	require.Equal(s.T(), s.patientID, *created.PatientID)
	require.Equal(s.T(), s.doctorID, *created.DoctorID)
	require.Equal(s.T(), s.unitID, *created.UnitID)
	require.Equal(s.T(), s.serviceID, *created.ServiceID)
}
