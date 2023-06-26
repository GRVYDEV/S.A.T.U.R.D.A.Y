import { generateRandomString, decodeDatachannelMessage } from "./util.js";

export class Client {
  constructor(stream, noPub, noSub, room, useDockerWs) {
    const configuration = {
      iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
    };

    this.stream = stream;
    this.noPub = noPub;
    this.noSub = noSub;
    this.room = room;
    this.useDockerWs = useDockerWs;

    this.muteBtn = document.getElementById("mic-mute");

    this.muteBtn.addEventListener("click", this.toggleMic);

    this.toggleMic();

    if (!noPub) {
      this.pub = new RTCPeerConnection(configuration);
      this.pubAns = false;
      this.pubCandidates = [];
      this.pub.onicecandidate = (e) => {
        const { candidate } = e;
        if (candidate) {
          console.log("[pub] ice candidate", JSON.stringify(candidate));
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

    if (!noSub) {
      this.sub = new RTCPeerConnection(configuration);
      this.subOff = false;
      this.subCandidates = [];
      this.sub.onicecandidate = (e) => {
        const { candidate } = e;
        if (candidate) {
          console.log("[sub] ice candidate", JSON.stringify(candidate));
          this.trickle(candidate, 1);
        }
      };

      this.sub.onconnectionstatechange = (e) => {
        const { connectionState } = this.sub;
        console.log("[sub] connstatechange", connectionState);
      };

      this.sub.ontrack = (e) => {
        console.log("houston we have a track", e);
        if (e.track.kind === "audio") {
          const audioEl = document.getElementById("saturday-audio");
          audioEl.srcObject = e.streams[0];
          console.log(e.streams[0].getAudioTracks());
        }
      };
      this.sub.ondatachannel = (e) => {
        const { channel } = e;
        console.log("got chan", channel);
        if (channel.label === "transcriptions") {
          channel.onmessage = (msg) => {
            decodeDatachannelMessage(msg.data);
          };
        }
      };
    }
  }

  toggleMic = () => {
    const audioTrack = this.stream.getAudioTracks()[0];
    audioTrack.enabled = !audioTrack.enabled;

    if (audioTrack.enabled) {
      this.muteBtn.innerText = "Mute Mic";
    } else {
      this.muteBtn.innerText = "Unmute Mic";
    }
  };

  async socketConnect() {
    return new Promise((resolve) => {
      let url = "ws://localhost:8088/ws";
      if (this.useDockerWs) {
        url = "ws://host.docker.internal:8088/ws";
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
        if (result) {
          if (result.type === "answer") {
            console.log("setting ans");
            await this.pub.setRemoteDescription(data.result);
            this.pubAns = true;
            this.pubCandidates.forEach((candidate) => {
              this.trickle(candidate, 0);
            });
          }
        } else {
          const { method, params } = data;
          if (method) {
            if (method === "trickle") {
              if (params.target === 0) {
                console.log("adding candidate for pub");
                await this.pub.addIceCandidate(params.candidate);
              }
              if (params.target === 1) {
                console.log("adding candidate for sub");
                if (!this.subOff) {
                  this.subCandidates.push(params.candidate);
                } else {
                  await this.sub.addIceCandidate(params.candidate);
                }
              }
            } else if (method === "offer") {
              console.log("setting offer");
              await this.sub.setRemoteDescription(params);
              const answer = await this.sub.createAnswer();
              await this.sub.setLocalDescription(answer);
              this.answer(answer);
              this.subOff = true;
              for (const candidate of this.subCandidates) {
                await this.sub.addIceCandidate(candidate);
              }
            }
          }
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
    await this.socketConnect();
    const join = {
      sid: this.room,
      uid: generateRandomString(10),
      config: {},
    };
    if (this.noSub) {
      join.config.NoSubscribe = true;
      join.config.NoAutoSubscribe = true;
    }
    if (this.noPub) {
      join.config.NoPublish = true;
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

    if (!this.noPub) {
      const offer = await this.pub.createOffer();
      await this.pub.setLocalDescription(offer);

      msg.params.offer = offer;
    }

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
  answer(answer) {
    console.log("answer", JSON.stringify(answer));
    this.socket.send(
      JSON.stringify({ method: "answer", params: { desc: answer } })
    );
  }
}
