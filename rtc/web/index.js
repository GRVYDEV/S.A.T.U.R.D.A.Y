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
  const queryString = window.location.search;
  const params = new URLSearchParams(queryString);
  const room = params.get("room") || "foo";
  const noSub = params.get("noSub") || false;
  const noPub = params.get("noPub") || false;
  const tester = params.get("tester") || false;
  console.log(noPub);
  let stream;
  if (!noPub) {
    stream = await getMedia();
    console.log(stream.getTracks());
    const videoEl = document.getElementById("local");
    videoEl.srcObject = stream;
    videoEl.muted = true;
  }

  const client = new Client(stream, noPub, noSub, room, tester);

  await client.join();
}

await init();
