// Package i18n centralizes the server messages that are visible to clients.
package i18n

import (
	"fmt"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	AuthUnknownProvider      = "auth.unknown_provider"
	AuthStartLoginFailed     = "auth.start_login_failed"
	AuthInvalidOAuthState    = "auth.invalid_oauth_state"
	AuthMissingCode          = "auth.missing_code"
	AuthCodeExchangeFailed   = "auth.code_exchange_failed"
	AuthProfileFetchFailed   = "auth.profile_fetch_failed"
	AuthProfilePersistFailed = "auth.profile_persist_failed"
	AuthSessionIssueFailed   = "auth.session_issue_failed"
	AuthUnauthorized         = "auth.unauthorized"
	AuthMissingCredentials   = "auth.missing_credentials"
	AuthInvalidToken         = "auth.invalid_token"

	PageRenderFailed = "page.render_failed"

	SimulationMethodNotAllowed   = "simulation.method_not_allowed"
	SimulationInvalidStartDate   = "simulation.invalid_start_date"
	SimulationInvalidEndDate     = "simulation.invalid_end_date"
	SimulationEndBeforeStart     = "simulation.end_before_start"
	SimulationRangeTooLarge      = "simulation.range_too_large"
	SimulationForecastTooFar     = "simulation.forecast_too_far"
	SimulationRangeSplitRequired = "simulation.range_split_required"
	SimulationWeatherProvider    = "simulation.weather_provider"
	SimulationNoWeatherData      = "simulation.no_weather_data"

	LoginTitle       = "login.title"
	LoginDescription = "login.description"
	LoginWithGoogle  = "login.with_google"
	LoginWithGitHub  = "login.with_github"

	DashboardTitle                = "dashboard.title"
	DashboardGreeting             = "dashboard.greeting"
	DashboardProfile              = "dashboard.profile"
	DashboardLogout               = "dashboard.logout"
	DashboardSystemStatus         = "dashboard.system_status"
	DashboardSimulation           = "dashboard.simulation"
	DashboardTotalLoss            = "dashboard.total_loss"
	DashboardVolume               = "dashboard.volume"
	DashboardLoweredMeters        = "dashboard.lowered_meters"
	DashboardLocation             = "dashboard.location"
	DashboardLocationHelp         = "dashboard.location_help"
	DashboardMapLabel             = "dashboard.map_label"
	DashboardUseMyLocation        = "dashboard.use_my_location"
	DashboardPlace                = "dashboard.place"
	DashboardLatitude             = "dashboard.latitude"
	DashboardLongitude            = "dashboard.longitude"
	DashboardWaterCharacteristics = "dashboard.water_characteristics"
	DashboardSurface              = "dashboard.surface"
	DashboardDepth                = "dashboard.depth"
	DashboardStartDate            = "dashboard.start_date"
	DashboardEndDate              = "dashboard.end_date"
	DashboardRunSimulation        = "dashboard.run_simulation"
	DashboardWeatherData          = "dashboard.weather_data"
	DashboardTemperature          = "dashboard.temperature"
	DashboardRelativeHumidity     = "dashboard.relative_humidity"
	DashboardWind                 = "dashboard.wind"
	DashboardRadiationEvaporation = "dashboard.radiation_evaporation"
	DashboardResult               = "dashboard.result"
	DashboardTotalEvaporation     = "dashboard.total_evaporation"
	DashboardEvaporatedVolume     = "dashboard.evaporated_volume"
	DashboardLinearMetersLowered  = "dashboard.linear_meters_lowered"
)

func init() {
	register(language.English, englishMessages)
	register(language.Spanish, spanishMessages)
	register(language.French, frenchMessages)
}

// Text formats a message key in the requested language.
func Text(tag language.Tag, key string, args ...any) string {
	return message.NewPrinter(tag).Sprintf(key, args...)
}

func register(tag language.Tag, messages map[string]string) {
	for key, text := range messages {
		if err := message.SetString(tag, key, text); err != nil {
			panic(fmt.Sprintf("registering %s message %q: %v", tag, key, err))
		}
	}
}

var englishMessages = map[string]string{
	AuthUnknownProvider: "unknown provider", AuthStartLoginFailed: "failed to start login", AuthInvalidOAuthState: "invalid OAuth state", AuthMissingCode: "missing authorization code", AuthCodeExchangeFailed: "failed to exchange authorization code", AuthProfileFetchFailed: "failed to fetch user profile", AuthProfilePersistFailed: "failed to persist user profile", AuthSessionIssueFailed: "failed to issue session token", AuthUnauthorized: "unauthorized", AuthMissingCredentials: "missing or malformed authorization", AuthInvalidToken: "invalid or expired token",
	PageRenderFailed:           "failed to render page",
	SimulationMethodNotAllowed: "method not allowed; use GET", SimulationInvalidStartDate: "invalid start date; expected YYYY-MM-DD", SimulationInvalidEndDate: "invalid end date; expected YYYY-MM-DD", SimulationEndBeforeStart: "end date must be after start date", SimulationRangeTooLarge: "date range cannot exceed %d days", SimulationForecastTooFar: "no forecast is available more than %d days from today", SimulationRangeSplitRequired: "date range is too wide; separate historical dates (more than %d days ago) from recent dates", SimulationWeatherProvider: "failed to contact the weather provider", SimulationNoWeatherData: "no weather data was found for the requested date range",
	LoginTitle: "Sign in - evap", LoginDescription: "Evaporation simulator for a body of water based on weather conditions", LoginWithGoogle: "Sign in with Google", LoginWithGitHub: "Sign in with GitHub",
	DashboardTitle: "Evaporation simulator", DashboardGreeting: "Hello, %s", DashboardProfile: "profile", DashboardLogout: "Sign out", DashboardSystemStatus: "System status", DashboardSimulation: "Simulation", DashboardTotalLoss: "Total loss", DashboardVolume: "Volume", DashboardLoweredMeters: "Lowered meters", DashboardLocation: "Location", DashboardLocationHelp: "Click the map, drag the marker, or use your current location.", DashboardMapLabel: "Map for selecting location", DashboardUseMyLocation: "Use my location", DashboardPlace: "Place", DashboardLatitude: "Latitude", DashboardLongitude: "Longitude", DashboardWaterCharacteristics: "Water body characteristics", DashboardSurface: "Surface (m²)", DashboardDepth: "Depth (m)", DashboardStartDate: "Start date", DashboardEndDate: "End date", DashboardRunSimulation: "Run simulation", DashboardWeatherData: "Weather data", DashboardTemperature: "Temperature (°C)", DashboardRelativeHumidity: "Relative humidity (%)", DashboardWind: "Wind (km/h)", DashboardRadiationEvaporation: "Solar radiation and evaporation", DashboardResult: "Result", DashboardTotalEvaporation: "Total evaporation", DashboardEvaporatedVolume: "Evaporated volume", DashboardLinearMetersLowered: "Linear meters lowered",
}

var spanishMessages = map[string]string{
	AuthUnknownProvider: "proveedor desconocido", AuthStartLoginFailed: "no se pudo iniciar sesión", AuthInvalidOAuthState: "estado OAuth inválido", AuthMissingCode: "falta el código de autorización", AuthCodeExchangeFailed: "no se pudo intercambiar el código de autorización", AuthProfileFetchFailed: "no se pudo obtener el perfil del usuario", AuthProfilePersistFailed: "no se pudo guardar el perfil del usuario", AuthSessionIssueFailed: "no se pudo emitir el token de sesión", AuthUnauthorized: "no autorizado", AuthMissingCredentials: "autorización faltante o malformada", AuthInvalidToken: "token inválido o expirado",
	PageRenderFailed:           "no se pudo renderizar la página",
	SimulationMethodNotAllowed: "método no permitido; use GET", SimulationInvalidStartDate: "fecha de inicio inválida; se espera AAAA-MM-DD", SimulationInvalidEndDate: "fecha de fin inválida; se espera AAAA-MM-DD", SimulationEndBeforeStart: "la fecha de fin debe ser posterior a la fecha de inicio", SimulationRangeTooLarge: "el rango de fechas no puede superar %d días", SimulationForecastTooFar: "no hay predicción disponible más allá de %d días desde hoy", SimulationRangeSplitRequired: "el rango de fechas es demasiado amplio; separe las fechas históricas (más de %d días) de las recientes", SimulationWeatherProvider: "no se pudo conectar con el proveedor meteorológico", SimulationNoWeatherData: "no se encontraron datos climáticos para el rango de fechas solicitado",
	LoginTitle: "Iniciar sesión - evap", LoginDescription: "Simulador de evaporación de una masa de agua en función del clima", LoginWithGoogle: "Iniciar sesión con Google", LoginWithGitHub: "Iniciar sesión con GitHub",
	DashboardTitle: "Simulador de evaporación", DashboardGreeting: "Hola, %s", DashboardProfile: "perfil", DashboardLogout: "Cerrar sesión", DashboardSystemStatus: "Estado del sistema", DashboardSimulation: "Simulación", DashboardTotalLoss: "Pérdida total", DashboardVolume: "Volumen", DashboardLoweredMeters: "Metros bajados", DashboardLocation: "Ubicación", DashboardLocationHelp: "Haz clic en el mapa, arrastra el marcador o utiliza tu ubicación actual.", DashboardMapLabel: "Mapa para seleccionar la ubicación", DashboardUseMyLocation: "Usar mi ubicación", DashboardPlace: "Lugar", DashboardLatitude: "Latitud", DashboardLongitude: "Longitud", DashboardWaterCharacteristics: "Características de la masa de agua", DashboardSurface: "Superficie (m²)", DashboardDepth: "Profundidad (m)", DashboardStartDate: "Fecha inicio", DashboardEndDate: "Fecha fin", DashboardRunSimulation: "Ejecutar simulación", DashboardWeatherData: "Datos meteorológicos", DashboardTemperature: "Temperatura (°C)", DashboardRelativeHumidity: "Humedad relativa (%)", DashboardWind: "Viento (km/h)", DashboardRadiationEvaporation: "Radiación solar y evaporación", DashboardResult: "Resultado", DashboardTotalEvaporation: "Evaporación total", DashboardEvaporatedVolume: "Volumen evaporado", DashboardLinearMetersLowered: "Metros lineales bajados",
}

var frenchMessages = map[string]string{
	AuthUnknownProvider: "fournisseur inconnu", AuthStartLoginFailed: "échec du démarrage de la connexion", AuthInvalidOAuthState: "état OAuth invalide", AuthMissingCode: "code d'autorisation manquant", AuthCodeExchangeFailed: "échec de l'échange du code d'autorisation", AuthProfileFetchFailed: "échec de récupération du profil utilisateur", AuthProfilePersistFailed: "échec de l'enregistrement du profil utilisateur", AuthSessionIssueFailed: "échec de l'émission du jeton de session", AuthUnauthorized: "non autorisé", AuthMissingCredentials: "autorisation manquante ou incorrecte", AuthInvalidToken: "jeton invalide ou expiré",
	PageRenderFailed:           "échec du rendu de la page",
	SimulationMethodNotAllowed: "méthode non autorisée ; utilisez GET", SimulationInvalidStartDate: "date de début invalide ; format attendu AAAA-MM-JJ", SimulationInvalidEndDate: "date de fin invalide ; format attendu AAAA-MM-JJ", SimulationEndBeforeStart: "la date de fin doit être postérieure à la date de début", SimulationRangeTooLarge: "la plage de dates ne peut pas dépasser %d jours", SimulationForecastTooFar: "aucune prévision n'est disponible au-delà de %d jours à partir d'aujourd'hui", SimulationRangeSplitRequired: "la plage de dates est trop étendue ; séparez les dates historiques (plus de %d jours) des dates récentes", SimulationWeatherProvider: "échec de connexion au fournisseur météorologique", SimulationNoWeatherData: "aucune donnée météorologique n'a été trouvée pour la plage de dates demandée",
	LoginTitle: "Se connecter - evap", LoginDescription: "Simulateur d'évaporation d'une masse d'eau selon les conditions météorologiques", LoginWithGoogle: "Se connecter avec Google", LoginWithGitHub: "Se connecter avec GitHub",
	DashboardTitle: "Simulateur d'évaporation", DashboardGreeting: "Bonjour, %s", DashboardProfile: "profil", DashboardLogout: "Se déconnecter", DashboardSystemStatus: "État du système", DashboardSimulation: "Simulation", DashboardTotalLoss: "Perte totale", DashboardVolume: "Volume", DashboardLoweredMeters: "Mètres abaissés", DashboardLocation: "Emplacement", DashboardLocationHelp: "Cliquez sur la carte, faites glisser le marqueur ou utilisez votre position actuelle.", DashboardMapLabel: "Carte pour sélectionner l'emplacement", DashboardUseMyLocation: "Utiliser ma position", DashboardPlace: "Lieu", DashboardLatitude: "Latitude", DashboardLongitude: "Longitude", DashboardWaterCharacteristics: "Caractéristiques de la masse d'eau", DashboardSurface: "Surface (m²)", DashboardDepth: "Profondeur (m)", DashboardStartDate: "Date de début", DashboardEndDate: "Date de fin", DashboardRunSimulation: "Exécuter la simulation", DashboardWeatherData: "Données météorologiques", DashboardTemperature: "Température (°C)", DashboardRelativeHumidity: "Humidité relative (%)", DashboardWind: "Vent (km/h)", DashboardRadiationEvaporation: "Rayonnement solaire et évaporation", DashboardResult: "Résultat", DashboardTotalEvaporation: "Évaporation totale", DashboardEvaporatedVolume: "Volume évaporé", DashboardLinearMetersLowered: "Mètres linéaires abaissés",
}
