# Deps

- mecab `brew install mecab`

- espeak `brew install espeak`

# Install steps

- Get python version > 3.7 < 3.11 with dev headers

- ensure `python-config --prefix` command works

- (optional) create a virtual env

- run `pip install -r requirements.txt`

- run `make run`

# Playing the audio file

The audio file generated is a float32 little endian binary pcm file with a sample rate of 22050. You can play it back with the command `❯ ffplay -f f32le -ar 22050 audio.pcm`
