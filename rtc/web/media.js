export async function getMedia() {
  try {
    return await navigator.mediaDevices.getUserMedia({
      audio: true,
      video: true,
    });
  } catch (err) {
    console.error("error getting media", err);
  }
}
