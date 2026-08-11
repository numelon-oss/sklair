const protocol = location.protocol === "https:" ? "wss" : "ws"; // useful if someone (incl me) puts sklair serve behind a tunnel temporarily
const ws = new WebSocket(`${protocol}://${location.host}/WEBSOCKET_PATH`);
ws.onmessage = (e) => {
	if (e.data === "reload") location.reload();
}