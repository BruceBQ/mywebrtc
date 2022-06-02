import rtc from "./rtc";

class Signaling {
    ws?: WebSocket;
    //   ws: WebSocket
    connect() {
        console.log("Start connnect http://localhost:3001/ws");
        this.ws = new WebSocket("ws://localhost:3001/ws");
        this.ws.onopen = () => {
            console.log("client side socket connection established");
            if (this.ws) {
                // this.ws?.send("ping")
                const data = { cameraId: "" };
                this.ws.send(JSON.stringify(data));
            }
        };

        this.ws.onclose = () => {
            console.log("client side socket connection disconnected");
        };
        this.ws.onerror = (error) => {
            console.log("Websocket error:", error);
            rtc.stop(true);
            alert(
                "Could not connect to websocket. Ready state: " +
                    (<WebSocket>error.target).readyState
            );
        };
        this.ws.onmessage = (message) => {};
    }

    close() {
        this.ws?.close();
    }
}

const signaling = new Signaling();
(<any>window).signaling = signaling;
export default signaling;
