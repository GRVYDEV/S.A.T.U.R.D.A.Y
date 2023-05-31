export async function getMedia() {
  try {
    return await navigator.mediaDevices.getUserMedia({
      audio: { noiseSuppression: true },
      video: true,
    });
  } catch (err) {
    console.error("error getting media", err.message);
  }
}

export async function getDevices() {
  await navigator.mediaDevices.getUserMedia({ audio: true, video: true });
  const devices = await navigator.mediaDevices.enumerateDevices();
  const videoDevices = devices.filter((d) => d.kind === "videoinput");
  const audioDevices = devices.filter((d) => d.kind === "audioinput");
  return { audioDevices, videoDevices };
}

export async function getCamera(deviceId) {
  let media;
  const videoConstraints = {
    deviceId: deviceId ? { exact: deviceId } : null,
  };

  return await navigator.mediaDevices.getUserMedia({
    video: videoConstraints,
    audio: false,
  });
}

export async function getMic(deviceId) {
  let media;
  const audioConstraints = {
    deviceId: deviceId ? { exact: deviceId } : null,
    noiseSuppression: true,
  };

  return await navigator.mediaDevices.getUserMedia({
    video: false,
    audio: audioConstraints,
  });
}
