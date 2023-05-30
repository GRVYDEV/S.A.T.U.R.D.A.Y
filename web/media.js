export async function getMedia() {
  try {
    return await navigator.mediaDevices.getUserMedia({
      audio: true,
      video: false,
    });
  } catch (err) {
    console.error("error getting media", err.message);
  }
}
