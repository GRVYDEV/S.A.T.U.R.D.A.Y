import { getMedia } from "./media.js";
import { Client } from "./client.js";

// pubConnection.onconnectionstatechange((conn) => {
//   console.log("[PUB] onStateChange", conn);
// });

// pubConnection.onicecandidate((candidate) => {
//   console.log("[PUB] onCandidate", candidate);
//   // handle me
// });

// subConnection.onconnectionstatechange((conn) => {
//   console.log("[SUB] onStateChange", conn);
// });

// subConnection.onicecandidate((candidate) => {
//   console.log("[SUB] onCandidate", candidate);
//   // handle me
// });

function handleJoin() {
  const join = {
    sid: "foo",
    uid: generateRandomString(10),
    offer: "foo",
  };
}

function handleOffer(offer) {}

async function init() {
  const stream = await getMedia();
  console.log(stream.getTracks());
  const videoEl = document.getElementById("local");
  videoEl.srcObject = stream;
  videoEl.muted = true;

  const client = new Client(stream);

  await client.join();
}

await init();
