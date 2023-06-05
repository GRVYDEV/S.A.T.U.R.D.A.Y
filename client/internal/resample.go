package internal

import "math"

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
