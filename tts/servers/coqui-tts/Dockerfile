ARG BASE=nvidia/cuda:11.8.0-base-ubuntu22.04
FROM ${BASE}
RUN apt-get update && apt-get upgrade -y
RUN apt-get install -y --no-install-recommends gcc g++ make python3 python3-dev python3-pip python3-venv python3-wheel espeak-ng libsndfile1-dev && rm -rf /var/lib/apt/lists/*
RUN pip3 install llvmlite --ignore-installed

# Set the working directory in the container
WORKDIR /app

# Copy the requirements file to the container
COPY ./tts/servers/coqui-tts/requirements.txt .
RUN pip install torch torchaudio --extra-index-url https://download.pytorch.org/whl/cu118
# Install the dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy the source code to the container
COPY ./tts/servers/coqui-tts/. .
RUN rm -rf /root/.cache/pip
RUN make install
# Expose the port on which the application will run (default is 8000)
EXPOSE 8000
# Start the application with Uvicorn
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]