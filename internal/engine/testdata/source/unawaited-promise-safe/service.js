const syncRemote = async () => {
  await fetch("https://example.com/health");
};

module.exports = {
  syncRemote,
};
