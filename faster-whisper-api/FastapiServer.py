import time
import numpy as np
from fastapi import FastAPI
from pydantic import BaseModel
from typing import List
from faster_whisper import WhisperModel

app = FastAPI()
model = WhisperModel("small", device="cuda", compute_type="int8")


class TranscriptionSegment(BaseModel):
    startTimestamp: int
    endTimestamp: int
    text: str


class Transcription(BaseModel):
    transcriptions: List[TranscriptionSegment]


def transform_segment(segment) -> dict:
    transcription_segment = {
        'startTimestamp': int(segment.start * 1000),
        'endTimestamp': int(segment.end * 1000),
        'text': segment.text
    }

    return transcription_segment


@app.post('/test/transcribe')
def transcribe(transcription_request: List[float]) -> Transcription:
    # Perform transcription on the audio data

    start = time.time()
    transcription = perform_transcription(transcription_request)
    end = time.time()

    print("Took:", end - start)
    print(transcription)
    return transcription


def perform_transcription(transcription_request):

    # Here you can implement your transcription logic
    # and generate the Transcription object
    segments, info = model.transcribe(np.array(transcription_request, dtype=np.float32),
                                      beam_size=5)

    return Transcription(transcriptions=[transform_segment(segment) for segment in segments])


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="localhost", port=8000)
