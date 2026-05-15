async function loadHealth() {
  return fetch("https://example.com/health");
}

module.exports = { loadHealth };
