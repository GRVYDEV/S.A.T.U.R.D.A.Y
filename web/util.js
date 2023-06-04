import { updateTranscriptions } from "./dom.js";

export function generateRandomString(length) {
  const characters =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  let result = "";
  for (let i = 0; i < length; i++) {
    result += characters.charAt(Math.floor(Math.random() * characters.length));
  }
  return result;
}

export function decodeDatachannelMessage(data) {
  const decoder = new TextDecoder();
  const arr = new Uint8Array(data);
  const json = JSON.parse(decoder.decode(arr));
  console.log("Got transcript:", json);
  updateTranscriptions(json.TranscribedText + json.CurrentTranscription);
}
