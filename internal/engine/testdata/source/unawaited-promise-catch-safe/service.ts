const syncRemote = async () => {
  client.request("/health").catch((err) => {
    console.error(err);
  });
};

export { syncRemote };
