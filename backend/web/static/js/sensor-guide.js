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
    } catch (error) {
      status.textContent = error.message;
    }
  });
})();