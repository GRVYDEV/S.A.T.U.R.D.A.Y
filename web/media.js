export async function getMedia() {
  let audio, video;

  try {
    const audioStream = await navigator.mediaDevices.getUserMedia({
      audio: { noiseSuppression: true },
    });
    audio = audioStream.getAudioTracks()[0];
  } catch (err) {
    console.error("error getting audio", err.message);
    throw new Error("audio is required to use Project S.A.T.U.R.D.A.Y");
  }
  try {
    const videoStream = await navigator.mediaDevices.getUserMedia({
      video: { height: { ideal: 1080 }, width: { ideal: 1920 } },
    });
    video = videoStream.getVideoTracks()[0];
  } catch (err) {
    console.error("error getting video", err.message);
  }
  const stream = new MediaStream([audio]);
  if (video) {
    stream.addTrack(video);
  }
  return stream;
}

export async function getDevices() {
  try {
    await navigator.mediaDevices.getUserMedia({ audio: true, video: true });
  } catch (err) {}
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
