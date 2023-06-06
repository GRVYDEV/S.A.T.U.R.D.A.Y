# STT

Tools for generating text from speech

## /engine

A go interface that allows for programatically generating text from audio

## /backends

A colletion of go backends that conform the the engine.Transcriber interface used to create text from audio with different inference mechanisms. PRs more than welcome for more backends :)
Backends are what actually does the audio -> text inference

## /servers

A collection of servers that can be used with various backends
