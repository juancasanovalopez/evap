package handlers

import (
	"context"
	"encoding/json"
	"evap-backend/internal/auth"
	"evap-backend/internal/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSimulateHappyPath(t *testing.T) {
	oldFetcher := WeatherForecastFetcher
	WeatherForecastFetcher = func(_ context.Context, lat, lon float64, start, end string) (OpenMeteoResponse, error) {
		resp := OpenMeteoResponse{}
		resp.Hourly.Time = []string{"2025-07-15T00:00", "2025-07-15T01:00"}
		resp.Hourly.Temperature2m = []float64{22, 23}
		resp.Hourly.RelativeHumidity2m = []float64{60, 62}
		resp.Hourly.WindSpeed10m = []float64{10, 11}
		resp.Hourly.ShortwaveRadiation = []float64{200, 220}
		return resp, nil
	}
	defer func() { WeatherForecastFetcher = oldFetcher }()

	issuer := auth.NewTokenIssuer("k", time.Hour)
	token, err := issuer.Issue("google#1", auth.Claims{Provider: "google", Email: "user@example.com"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/simulate?area=32&profundidad=1.2&lat=40.4167&lon=-3.7037&fecha_inicio=2025-07-15&fecha_fin=2025-07-17", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler := middleware.JWTAuth(issuer)(http.HandlerFunc(Simulate))
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	require.Contains(t, result, "litros_totales_perdidos")
	require.Contains(t, result, "reporte_horario")
}
