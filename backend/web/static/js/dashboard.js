// Dashboard: mapa de ubicación (Leaflet), geolocalización, geocoding inverso
// (Nominatim) y simulación real contra /api/v1/simulate con gráficas (Chart.js).

(function () {
  const DEFAULT_LAT = 40.4167;
  const DEFAULT_LON = -3.7037;

  const latInput = document.getElementById('lat');
  const lonInput = document.getElementById('lon');
  const placeInput = document.getElementById('place-name');
  const locationStatus = document.getElementById('location-status');
  const simulationStatus = document.getElementById('simulation-status');
  const form = document.getElementById('simulation-form');
  const fechaInicioInput = document.getElementById('fecha_inicio');
  const fechaFinInput = document.getElementById('fecha_fin');
  const logoutBtn = document.getElementById('logout-btn');

  function toISODate(date) {
    return date.toISOString().slice(0, 10);
  }

  // El proveedor meteorológico sólo sirve un rango limitado alrededor de hoy.
  (function initDates() {
    const today = new Date();
    const end = new Date(today);
    end.setDate(end.getDate() + 2);
    if (!fechaInicioInput.value) fechaInicioInput.value = toISODate(today);
    if (!fechaFinInput.value) fechaFinInput.value = toISODate(end);
  })();

  if (logoutBtn) {
    logoutBtn.addEventListener('click', async () => {
      logoutBtn.disabled = true;
      try {
        await fetch('/auth/logout', { method: 'POST', credentials: 'include' });
      } finally {
        window.location.href = '/login';
      }
    });
  }

  let reverseGeocodeTimer = null;

  const map = L.map('map').setView([DEFAULT_LAT, DEFAULT_LON], 11);
  L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; OpenStreetMap contributors',
  }).addTo(map);

  const marker = L.marker([DEFAULT_LAT, DEFAULT_LON], { draggable: true }).addTo(map);

  function setCoordinates(lat, lon) {
    latInput.value = lat.toFixed(4);
    lonInput.value = lon.toFixed(4);
  }

  function reverseGeocode(lat, lon) {
    clearTimeout(reverseGeocodeTimer);
    reverseGeocodeTimer = setTimeout(async () => {
      try {
        const url = `https://nominatim.openstreetmap.org/reverse?format=jsonv2&lat=${lat}&lon=${lon}`;
        const res = await fetch(url, { headers: { Accept: 'application/json' } });
        if (!res.ok) throw new Error('reverse geocode failed');
        const data = await res.json();
        placeInput.value = data.display_name || `${lat.toFixed(4)}, ${lon.toFixed(4)}`;
      } catch (_err) {
        placeInput.value = `${lat.toFixed(4)}, ${lon.toFixed(4)}`;
      }
    }, 500);
  }

  function onLocationChange(lat, lon) {
    setCoordinates(lat, lon);
    reverseGeocode(lat, lon);
  }

  marker.on('dragend', () => {
    const pos = marker.getLatLng();
    onLocationChange(pos.lat, pos.lng);
  });

  map.on('click', (e) => {
    marker.setLatLng(e.latlng);
    onLocationChange(e.latlng.lat, e.latlng.lng);
  });

  const geolocateBtn = document.getElementById('geolocate-btn');
  geolocateBtn.addEventListener('click', () => {
    if (!('geolocation' in navigator)) {
      locationStatus.textContent = 'Geolocalización no disponible en este navegador.';
      return;
    }
    locationStatus.textContent = 'Buscando tu ubicación…';
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        const { latitude, longitude } = pos.coords;
        map.setView([latitude, longitude], 13);
        marker.setLatLng([latitude, longitude]);
        onLocationChange(latitude, longitude);
        locationStatus.textContent = '';
      },
      () => {
        locationStatus.textContent = 'No se pudo obtener tu ubicación. Usa el mapa manualmente.';
      },
    );
  });

  reverseGeocode(DEFAULT_LAT, DEFAULT_LON);

  const chartColors = {
    temperature: '#1f6feb',
    water: '#0e7c61',
    humidity: '#1f6feb',
    wind: '#1f6feb',
    radiation: '#f59e0b',
    evaporation: '#0e7c61',
  };

  function makeLineChart(canvasId, datasets, labels, extraScales) {
    const ctx = document.getElementById(canvasId);
    if (!ctx) return null;
    return new Chart(ctx, {
      type: 'line',
      data: { labels, datasets },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: { mode: 'index', intersect: false },
        scales: Object.assign({ x: { ticks: { maxTicksLimit: 8 } } }, extraScales || {}),
      },
    });
  }

  let charts = {};

  function destroyCharts() {
    Object.values(charts).forEach((c) => c && c.destroy());
    charts = {};
  }

  function renderCharts(reporte) {
    destroyCharts();
    if (!Array.isArray(reporte) || reporte.length === 0) return;
    const labels = reporte.map((h) => h.timestamp);

    charts.temperature = makeLineChart(
      'chart-temperature',
      [
        { label: 'Aire', data: reporte.map((h) => h.temperatura_aire), borderColor: chartColors.temperature, tension: 0.3 },
        { label: 'Agua', data: reporte.map((h) => h.t_agua_calculada), borderColor: chartColors.water, tension: 0.3 },
      ],
      labels,
    );

    charts.humidity = makeLineChart(
      'chart-humidity',
      [{ label: 'Humedad relativa', data: reporte.map((h) => h.humedad_relativa), borderColor: chartColors.humidity, tension: 0.3 }],
      labels,
    );

    charts.wind = makeLineChart(
      'chart-wind',
      [{ label: 'Viento', data: reporte.map((h) => h.viento_kmh), borderColor: chartColors.wind, tension: 0.3 }],
      labels,
    );

    charts.radiation = makeLineChart(
      'chart-radiation',
      [
        { label: 'Radiación solar (W/m²)', data: reporte.map((h) => h.radiacion_solar), borderColor: chartColors.radiation, tension: 0.3 },
        { label: 'Evaporación (L/h)', data: reporte.map((h) => h.evap_litros_hora), borderColor: chartColors.evaporation, tension: 0.3, yAxisID: 'y1' },
      ],
      labels,
      {
        y: { type: 'linear', position: 'left' },
        y1: { type: 'linear', position: 'right', grid: { drawOnChartArea: false } },
      },
    );
  }

  function setResult(id, value) {
    document.getElementById(id).textContent = value;
  }

  async function runSimulation() {
    const area = document.getElementById('area').value;
    const profundidad = document.getElementById('profundidad').value;
    const lat = latInput.value;
    const lon = lonInput.value;
    const fechaInicio = fechaInicioInput.value;
    const fechaFin = fechaFinInput.value;

    const params = new URLSearchParams({
      area, profundidad, lat, lon, fecha_inicio: fechaInicio, fecha_fin: fechaFin,
    });

    simulationStatus.textContent = 'Calculando…';
    try {
      const res = await fetch(`/api/v1/simulate?${params.toString()}`, { credentials: 'include' });
      if (res.status === 401) {
        window.location.href = '/login';
        return;
      }
      const data = await res.json();
      if (!res.ok) {
        destroyCharts();
        simulationStatus.textContent = data.error || 'No se pudo ejecutar la simulación.';
        return;
      }

      simulationStatus.textContent = '';
      setResult('total-mm', `${data.mm_totales_descendidos.toFixed(2)} mm`);
      setResult('total-liters', `${data.litros_totales_perdidos.toFixed(1)} L`);
      setResult('total-meters', `${data.metros_lineales_bajo.toFixed(5)} m`);
      setResult('hero-daily-mm', `${data.mm_totales_descendidos.toFixed(2)} mm`);
      setResult('hero-liters', `${data.litros_totales_perdidos.toFixed(1)} L`);
      setResult('hero-area', `${data.metros_lineales_bajo.toFixed(5)} m`);

      renderCharts(data.reporte_horario);
    } catch (_err) {
      simulationStatus.textContent = 'Fallo de red al contactar con el servidor.';
    }
  }

  form.addEventListener('submit', (e) => {
    e.preventDefault();
    runSimulation();
  });

  runSimulation();
})();
