package util

import (
	"encoding/binary"
	"math"
)

func BinaryToFloat32(data []byte) []float32 {
	floats := make([]float32, 0, len(data)/4)
	for {
		if len(data) == 0 {
			break
		}

		u := binary.LittleEndian.Uint32(data[:4])
		f := math.Float32frombits(u)
		floats = append(floats, f)
		data = data[4:]
	}
	return floats
}

func ConvertToDualChannel(input []float32) []float32 {
	// Calculate the number of samples in the input audio
	numSamples := len(input)

	// Create a new slice for the dual-channel output audio
	output := make([]float32, numSamples*2)

	// Iterate over each sample in the input audio
	for i := 0; i < numSamples; i++ {
		// Get the current sample value
		sample := input[i]

		// Set the same sample value for both channels in the output audio
		output[i*2] = sample   // Left channel
		output[i*2+1] = sample // Right channel
	}

	return output
}
