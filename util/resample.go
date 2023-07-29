package util

import "math"

// Resample takes an input signal and resamples it to a different length using cubic interpolation.
func Resample(input []float32, inputSampleRate, targetSampleRate int) []float32 {
	inputLength := len(input)
	outputLength := int(math.Round(float64(inputLength) * float64(float64(targetSampleRate)/float64(inputSampleRate))))

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
