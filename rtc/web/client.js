import { generateRandomString } from "./util.js";

export class Client {
  constructor(stream) {
    const configuration = {
      iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
    };

    this.stream = stream;

    this.sub = new RTCPeerConnection(configuration);
    this.pub = new RTCPeerConnection(configuration);
    this.pubAns = false;
    this.pubCandidates = [];

    this.pub.onicecandidate = (e) => {
      const { candidate } = e;
      if (candidate) {
        console.log("[pub] ice candidate");
        if (this.pubAns) {
          this.trickle(candidate, 0);
        } else {
          this.pubCandidates.push(candidate);
        }
      }
    };

    this.pub.onconnectionstatechange = (e) => {
      const { connectionState } = this.pub;
      console.log("[pub] connstatechange", connectionState);
    };
  }

  async socketConnect(tester = false) {
    return new Promise((resolve) => {
      let url = "ws://localhost:8080/ws";
      if (tester) {
        url = "ws://host.docker.internal:8080/ws";
      }
      this.socket = new WebSocket(url);

      // Event listener for when the WebSocket connection is opened
      this.socket.addEventListener("open", (event) => {
        console.log("WebSocket connection opened");
        resolve();
      });

      // Event listener for when a message is received over the WebSocket
      this.socket.addEventListener("message", async (event) => {
        const data = JSON.parse(event.data);
        console.log(`WebSocket message received: ${data}`, data);
        const { result } = data;
        if (result && result.type === "answer") {
          console.log("setting ans");
          await this.pub.setRemoteDescription(data.result);
          this.pubAns = true;
          this.pubCandidates.forEach((candidate) => {
            this.trickle(candidate, 0);
          });
        }
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
    const queryString = window.location.search;
    const params = new URLSearchParams(queryString);
    const room = params.get("room") || "foo";
    const noSub = params.get("noSub") || false;
    await this.socketConnect();
    const join = {
      sid: room,
      uid: generateRandomString(10),
    };
    if (noSub) {
      join.config = {
        NoSubscribe: true,
        NoAutoSubscribe: true,
      };
    }
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
