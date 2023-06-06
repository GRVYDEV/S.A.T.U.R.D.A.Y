import sys
import base64
import numpy as np
from fastapi import FastAPI
from TTS.api import TTS
from pydantic import BaseModel

app = FastAPI()


class SynthesizeResponse(BaseModel):
    sample_rate: int
    data: bytes


class SynthesizeRequest(BaseModel):
    text: str


# output will be signed int32 little endian
# NOTE endianess in python depends on proc. you can check with sys.byteorder
tts = TTS("tts_models/en/vctk/vits")
speaker = "p335"
sample_rate = tts.synthesizer.output_sample_rate


# synthesize will run tts on the provided string and output float32 little endian bytes
def run_synthesis(text) -> bytes:
    pcm = tts.tts(text, speaker=speaker)
    arr = np.array(pcm, dtype=np.float32)
    if sys.byteorder == "big":
        arr.byteswap()

    arr.tofile("audio.pcm")
    return arr.tobytes()


@app.post("/synthesize")
async def synthesize(request: SynthesizeRequest) -> SynthesizeResponse:
    print("running synth with text: ", request.text)
    pcm = run_synthesis(request.text)

    return SynthesizeResponse(data=base64.b64encode(pcm), sample_rate=sample_rate)
