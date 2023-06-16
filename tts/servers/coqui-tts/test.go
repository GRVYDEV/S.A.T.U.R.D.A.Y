package main

import (
	"encoding/binary"
	"log"
	"math"
	"os"
)

func main() {
	data, err := os.ReadFile("./audio.pcm")
	if err != nil {
		log.Fatalf("error reading file %+v", err)
	}

	floats := BinaryToFloat32(data)

	res := CubicSplineInterpolation(floats, 22050, 48000)

	f, err := os.OpenFile("audio_test.pcm", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = binary.Write(f, binary.LittleEndian, res)
	if err != nil {
		panic(err)
	}

}

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

// CubicSplineInterpolation resamples the given audio data using cubic spline interpolation.
func CubicSplineInterpolation(input []float32, inputSampleRate, targetSampleRate int) []float32 {
	inputLength := len(input)
	outputLength := inputLength * (targetSampleRate / inputSampleRate)

	output := make([]float32, outputLength)

	step := float64(inputLength-1) / float64(outputLength-1)

	for i := 0; i < outputLength; i++ {
		index := float64(i) * step
		lower := int(math.Floor(index))
		upper := int(math.Ceil(index))

		if lower >= inputLength-1 {
			lower = inputLength - 2
		}

		// Ensure the upper index doesn't go out of bounds
		if upper >= inputLength-1 {
			upper = inputLength - 2
		}

		// Calculate the fractional part
		frac := float32(index - float64(lower))

		// Calculate the coefficients for the cubic spline interpolation
		a := -0.5*input[lower] + 1.5*input[lower+1] - 1.5*input[upper] + 0.5*input[upper+1]
		b := input[lower] - 2.5*input[lower+1] + 2*input[upper] - 0.5*input[upper+1]
		c := -0.5*input[lower] + 0.5*input[upper]
		d := input[lower+1]

		// Perform cubic spline interpolation
		output[i] = float32(((a*frac+b)*frac+c)*frac + d)
	}

	return output
}

// Resample takes an input signal and resamples it to a different length using cubic interpolation.
func Resample(input []float32, inputSampleRate, targetSampleRate int) []float32 {
	inputLength := len(input)
	outputLength := inputLength * (targetSampleRate / inputSampleRate)

	output := make([]float32, outputLength)

	for i := 0; i < outputLength; i++ {
		index := float32(i) * float32(inputLength-1) / float32(outputLength-1)
		leftIndex := int(math.Floor(float64(index)))

		weights := make([]float32, 4)
		for j := 0; j < 4; j++ {
			d := index - float32(leftIndex+j)
			weights[j] = cubicWeight(d)
		}

		for j := 0; j < 4; j++ {
			inputIndex := leftIndex + j - 1
			if inputIndex >= 0 && inputIndex < inputLength {
				output[i] += input[inputIndex] * weights[j]
			}
		}
	}

	// Normalize the output
	maxSample := getMaxSample(output)
	output = normalize(output, maxSample)

	return output
}

// cubicWeight calculates the weight for the cubic interpolation.
func cubicWeight(d float32) float32 {
	if d < 0 {
		d = -d
	}

	if d < 1 {
		return (1.5*d-2.5)*d*d + 1
	} else if d < 2 {
		d -= 2
		return ((-0.5*d+2.5)*d-4)*d + 2
	}

	return 0
}

// getMaxSample returns the maximum absolute value in the given samples.
func getMaxSample(samples []float32) float32 {
	max := float32(0)
	for _, sample := range samples {
		absSample := float32(math.Abs(float64(sample)))
		if absSample > max {
			max = absSample
		}
	}
	return max
}

// normalize scales the samples by a given factor to ensure they fall within the desired range.
func normalize(samples []float32, factor float32) []float32 {
	normalized := make([]float32, len(samples))
	for i, sample := range samples {
		normalized[i] = sample / factor
	}
	return normalized
}
