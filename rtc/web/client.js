import { generateRandomString } from "./util.js";

export class Client {
  constructor(stream) {
    const configuration = {
      iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
    };

    this.stream = stream;

    this.sub = new RTCPeerConnection(configuration);
    this.pub = new RTCPeerConnection(configuration);

    this.pub.onicecandidate = (e) => {
      const { candidate } = e;
      if (candidate) {
        console.log("[pub] ice candidate");
        this.trickle(candidate, 0);
      }
    };

    this.pub.onconnectionstatechange = (e) => {
      const { connectionState } = this.pub;
      console.log("[pub] connstatechange", connectionState);
    };
  }

  async socketConnect() {
    return new Promise((resolve) => {
      this.socket = new WebSocket("ws://localhost:8080/ws");

      // Event listener for when the WebSocket connection is opened
      this.socket.addEventListener("open", (event) => {
        console.log("WebSocket connection opened");
        resolve();
      });

      // Event listener for when a message is received over the WebSocket
      this.socket.addEventListener("message", (event) => {
        console.log(`WebSocket message received: ${event.data}`);
      });

      // Event listener for when the WebSocket connection is closed
      this.socket.addEventListener("close", (event) => {
        console.log("WebSocket connection closed");
      });

      // Event listener for errors that occur on the WebSocket
      this.socket.addEventListener("error", (event) => {
        console.log(`WebSocket error: ${event}`);
      });
    });
  }

  async join() {
    await this.socketConnect();
    const join = {
      sid: "foo",
      uid: generateRandomString(10),
    };
    const msg = {
      method: "join",
      params: join,
    };

    if (this.stream) {
      this.stream.getTracks().forEach((track) => {
        this.pub.addTransceiver(track);
      });
    }

    const offer = await this.pub.createOffer();
    await this.pub.setLocalDescription(offer);

    msg.params.offer = offer;

    this.socket.send(JSON.stringify(msg));
  }

  trickle(candidate, target) {
    const msg = {
      method: "trickle",
      params: {
        target,
        candidate,
      },
    };

    this.socket.send(JSON.stringify(msg));
  }
}
