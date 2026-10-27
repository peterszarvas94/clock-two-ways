const face = document.getElementById("face");
const zone = face.dataset.zone;
const epoch = Number(face.dataset.epoch);
const started = performance.now();
const partsFormat = new Intl.DateTimeFormat("en-GB", {
  timeZone: zone,
  hour: "2-digit",
  minute: "2-digit",
  second: "2-digit",
  hourCycle: "h23",
});
const dateFormat = new Intl.DateTimeFormat("en-GB", {
  timeZone: zone,
  weekday: "short",
  day: "2-digit",
  month: "short",
  year: "numeric",
});
// Keep ticking from the server's initial time, without further requests.
function tick() {
  const now = new Date(epoch + performance.now() - started);
  const parts = Object.fromEntries(
    partsFormat.formatToParts(now).map((p) => [p.type, p.value]),
  );
  face.style.setProperty("--hour", Number(parts.hour) % 12);
  face.style.setProperty("--minute", Number(parts.minute));
  face.style.setProperty("--second", Number(parts.second));
  document.getElementById("digital").textContent = partsFormat.format(now);
  document.getElementById("date").textContent = dateFormat.format(now);
}
setInterval(tick, 1000);
document.addEventListener("visibilitychange", () => {
  if (!document.hidden) tick();
});
