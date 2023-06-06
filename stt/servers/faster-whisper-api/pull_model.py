from faster_whisper import WhisperModel

model = WhisperModel("small", device="cpu", compute_type="int8")
