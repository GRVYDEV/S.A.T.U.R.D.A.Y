import { getDevices } from "./media.js";

export function updateTranscriptions(text) {
  const div = document.getElementById("transcriptions");
  div.innerHTML = text;
}

export async function initializeDeviceSelect() {
  const videoSelect = document.getElementById("camera");
  const audioSelect = document.getElementById("mic");

  const { videoDevices, audioDevices } = await getDevices();

  videoSelect.disabled = false;
  videoDevices.forEach((device, index) => {
    videoSelect.options[index] = new Option(device.label, device.deviceId);
  });

  audioSelect.disabled = false;
  audioDevices.forEach((device, index) => {
    audioSelect.options[index] = new Option(device.label, device.deviceId);
  });
}
