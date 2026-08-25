package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"evap-backend/internal/middleware"
)

const (
	albedo       = 0.15
	rhoAgua      = 1000.0
	cpAgua       = 4186.0
	sigma        = 5.67e-8
	tAguaInicial = 20.0
)

// OpenMeteoResponse maps the Open-Meteo hourly forecast response.
type OpenMeteoResponse struct {
	Hourly struct {
		Time               []string  `json:"time"`
		Temperature2m      []float64 `json:"temperature_2m"`
		RelativeHumidity2m []float64 `json:"relative_humidity_2m"`
		WindSpeed10m       []float64 `json:"wind_speed_10m"`
		ShortwaveRadiation []float64 `json:"shortwave_radiation"`
	} `json:"hourly"`
}

// SimulacionResult is the public JSON payload returned to the authenticated user.
type SimulacionResult struct {
	LitrosTotalesPerdidos float64        `json:"litros_totales_perdidos"`
	MmTotalesDescendidos  float64        `json:"mm_totales_descendidos"`
	MetrosLinealesBajo    float64        `json:"metros_lineales_bajo"`
	ReporteHorario        []HoraDataCell `json:"reporte_horario"`
}

// HoraDataCell holds per-hour calculation details.
type HoraDataCell struct {
	HoraIndex      int     `json:"hora_index"`
	Timestamp      string  `json:"timestamp"`
	TaguaCalculada float64 `json:"t_agua_calculada"`
	RadiacionSolar float64 `json:"radiacion_solar"`
	EvapLitrosHora float64 `json:"evap_litros_hora"`
	EvapMmHora     float64 `json:"evap_mm_hora"`
}

// WeatherForecastFetcher is injected so tests can stub the provider call.
var WeatherForecastFetcher = func(ctx context.Context, lat, lon float64, startDate, endDate string) (OpenMeteoResponse, error) {
	params := url.Values{}
	params.Set("latitude", strconv.FormatFloat(lat, 'f', -1, 64))
	params.Set("longitude", strconv.FormatFloat(lon, 'f', -1, 64))
	params.Set("hourly", "temperature_2m,relative_humidity_2m,wind_speed_10m,shortwave_radiation")
	params.Set("start_date", startDate)
	params.Set("end_date", endDate)
	params.Set("timezone", "auto")

	apiURL := "https://api.open-meteo.com/v1/forecast?" + params.Encode()
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return OpenMeteoResponse{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return OpenMeteoResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return OpenMeteoResponse{}, fmt.Errorf("meteorological provider returned %d", resp.StatusCode)
	}

	var data OpenMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return OpenMeteoResponse{}, err
	}
	return data, nil
}

// Simulate runs the thermal evaporation model for an authenticated session.
func Simulate(w http.ResponseWriter, r *http.Request) {
	if _, ok := middleware.ClaimsFromContext(r.Context()); !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "Método no permitido. Utilizar GET"}`, http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query()
	area := parseFloatOrDefault(q.Get("area"), 32.0)
	profundidad := parseFloatOrDefault(q.Get("profundidad"), 1.2)
	lat := parseFloatOrDefault(q.Get("lat"), 40.4167)
	lon := parseFloatOrDefault(q.Get("lon"), -3.7037)
	fechaInicio := q.Get("fecha_inicio")
	if fechaInicio == "" {
		fechaInicio = "2025-07-15"
	}
	fechaFin := q.Get("fecha_fin")
	if fechaFin == "" {
		fechaFin = "2025-07-17"
	}

	climaData, err := WeatherForecastFetcher(r.Context(), lat, lon, fechaInicio, fechaFin)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Fallo al conectar con el proveedor meteorológico: " + err.Error()})
		return
	}

	pasos := len(climaData.Hourly.Temperature2m)
	if pasos == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "No se encontraron datos climáticos para el rango de fechas proporcionado."})
		return
	}

	dt := 3600.0
	tAguaActual := tAguaInicial
	masaAgua := area * profundidad * rhoAgua

	var reporteHorario []HoraDataCell
	var litrosTotalesPerdidos float64
	var mmTotalesDescendidos float64

	for p := 0; p < pasos; p++ {
		gSolar := climaData.Hourly.ShortwaveRadiation[p]
		tAireInst := climaData.Hourly.Temperature2m[p]
		hrInst := climaData.Hourly.RelativeHumidity2m[p]
		vientoInst := climaData.Hourly.WindSpeed10m[p] / 3.6

		pSatAgua := 0.61078 * math.Exp((17.27*tAguaActual)/(tAguaActual+237.3))
		pSatAire := 0.61078 * math.Exp((17.27*tAireInst)/(tAireInst+237.3))
		pParcialAire := pSatAire * (hrInst / 100.0)

		lv := (2501.0 - 2.36*tAguaActual) * 1000.0
		tasaEvapM2S := ((0.0887 + 0.0781*vientoInst) / 3600.0) * math.Max(0.0, pSatAgua-pParcialAire)
		qEvap := tasaEvapM2S * lv

		hc := 5.7 + 3.8*vientoInst
		qConv := hc * (tAguaActual - tAireInst)

		tAireK := tAireInst + 273.15
		tCieloK := 0.0552 * math.Pow(tAireK, 1.5)
		tAguaK := tAguaActual + 273.15
		qRadNeta := sigma * 0.95 * (math.Pow(tAguaK, 4) - math.Pow(tCieloK, 4))
		qSolarAbsorbida := gSolar * (1.0 - albedo)
		qNeto := qSolarAbsorbida - qEvap - qConv - qRadNeta
		potenciaNetaTotal := qNeto * area

		dTAgua := (potenciaNetaTotal / (masaAgua * cpAgua)) * dt
		tAguaActual += dTAgua

		evapLitrosHora := tasaEvapM2S * area * 3600.0
		evapMmHora := evapLitrosHora / area

		litrosTotalesPerdidos += evapLitrosHora
		mmTotalesDescendidos += evapMmHora

		reporteHorario = append(reporteHorario, HoraDataCell{
			HoraIndex:      p + 1,
			Timestamp:      climaData.Hourly.Time[p],
			TaguaCalculada: math.Round(tAguaActual*100) / 100,
			RadiacionSolar: gSolar,
			EvapLitrosHora: math.Round(evapLitrosHora*100) / 100,
			EvapMmHora:     math.Round(evapMmHora*1000) / 1000,
		})
	}

	resultado := SimulacionResult{
		LitrosTotalesPerdidos: math.Round(litrosTotalesPerdidos*10) / 10,
		MmTotalesDescendidos:  math.Round(mmTotalesDescendidos*100) / 100,
		MetrosLinealesBajo:    (math.Round(mmTotalesDescendidos*100) / 100) / 1000.0,
		ReporteHorario:        reporteHorario,
	}

	writeJSON(w, http.StatusOK, resultado)
}

func parseFloatOrDefault(val string, def float64) float64 {
	if val == "" {
		return def
	}
	parsed, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return def
	}
	return parsed
}
