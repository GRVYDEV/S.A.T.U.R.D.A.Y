# TTS

Tools for generating speech from text

## /engine

A go interface that allows for programatically generating audio from text

## /backends

A colletion of go backends that conform the the engine.Synthesizer interface used to create audio from text with different inference mechanisms. PRs more than welcome for more backends :)
Backends are what actually does the text -> audio inference

## /servers

A collection of servers that can be used with various backends
