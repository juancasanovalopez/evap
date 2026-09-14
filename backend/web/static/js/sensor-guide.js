(() => {
  const form = document.querySelector("#sensor-form");
  if (!form) return;
  const status = document.querySelector("#sensor-status");
  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    status.textContent = "Creando sensor...";
    const name = new FormData(form).get("name");
    try {
      const response = await fetch("/api/v1/sensors", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name }),
      });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || "No se pudo crear el sensor.");
      document.querySelector("#sensor-endpoint").textContent = data.endpoint;
      document.querySelector("#sensor-topic").textContent = data.topic;
      document.querySelector("#created-thing").textContent = data.thing_name;
      document.querySelector("#certificate-pem").value = data.certificate_pem;
      document.querySelector("#private-key").value = data.private_key;
      document.querySelector("#sensor-credentials").hidden = false;
      status.textContent = "Sensor creado. Guarda las credenciales antes de cerrar esta página.";
      form.reset();
      loadSensors();
    } catch (error) {
      status.textContent = error.message;
    }
  });

  function escapeHtml(value) {
    return String(value).replace(/[&<>"']/g, (c) => ({
      "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;",
    }[c]));
  }

  const sensorsStatus = document.querySelector("#sensors-status");
  const sensorsList = document.querySelector("#sensors-list");

  async function loadSensors() {
    if (!sensorsList) return;
    try {
      const res = await fetch("/api/v1/sensors", { credentials: "include" });
      if (res.status === 401) {
        window.location.href = "/login";
        return;
      }
      const data = await res.json();
      if (!res.ok) {
        sensorsStatus.textContent = data.error || "No se pudieron cargar los sensores.";
        return;
      }
      sensorsStatus.textContent = "";
      sensorsList.innerHTML = "";
      if (!Array.isArray(data) || data.length === 0) {
        sensorsStatus.textContent = "Todavía no has añadido ningún sensor.";
        return;
      }
      for (const sensor of data) {
        const li = document.createElement("li");
        li.textContent = sensor.thing_name;
        sensorsList.appendChild(li);
      }
    } catch (_err) {
      sensorsStatus.textContent = "Fallo de red al contactar con el servidor.";
    }
  }

  const messagesStatus = document.querySelector("#messages-status");
  const messagesBody = document.querySelector("#messages-body");

  async function loadMessages() {
    if (!messagesBody) return;
    try {
      const res = await fetch("/api/v1/readings?limit=50", { credentials: "include" });
      if (res.status === 401) {
        window.location.href = "/login";
        return;
      }
      const data = await res.json();
      if (!res.ok) {
        messagesStatus.textContent = data.error || "No se pudieron cargar los mensajes recibidos.";
        return;
      }
      messagesStatus.textContent = "";
      messagesBody.innerHTML = "";
      if (!Array.isArray(data) || data.length === 0) {
        messagesStatus.textContent = "No se han recibido mensajes en las últimas 24 horas.";
        return;
      }
      for (const reading of data) {
        const row = document.createElement("tr");
        row.innerHTML = `<td>${escapeHtml(reading.device_id)}</td><td>${escapeHtml(reading.temperature)}</td><td>${escapeHtml(reading.timestamp)}</td>`;
        messagesBody.appendChild(row);
      }
    } catch (_err) {
      messagesStatus.textContent = "Fallo de red al contactar con el servidor.";
    }
  }

  loadSensors();
  loadMessages();
  setInterval(loadMessages, 30000);
})();