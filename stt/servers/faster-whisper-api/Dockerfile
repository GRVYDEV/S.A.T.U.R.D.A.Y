# Use the official Python base image
FROM python:3.10-slim

# Set the working directory in the container
WORKDIR /app

# Copy the requirements file to the container
COPY ./stt/servers/faster-whisper-api/requirements.txt .

# Install the dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy the source code to the container
COPY ./stt/servers/faster-whisper-api/. .

# Expose the port on which the application will run (default is 8000)
EXPOSE 8000
RUN python pull_model.py
# Start the application with Uvicorn
CMD ["uvicorn", "FastapiServer:app", "--host", "0.0.0.0", "--port", "8000"]