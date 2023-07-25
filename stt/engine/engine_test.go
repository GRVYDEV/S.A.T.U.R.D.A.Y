package engine

import "testing"

func TestVAD(t *testing.T) {
	// Define test cases with input frames and expected output
	testCases := []struct {
		frame          []float32
		expectedOutput bool
	}{
		// Test case 1: Valid speech frame
		{
			frame:          []float32{0.1, 0.2, 0.3, 0.2, 0.1, 0.1, 0.2, 0.3, 0.2, 0.1, 0.1, 0.2, 0.3, 0.2, 0.1},
			expectedOutput: true,
		},
		// Test case 2: Silent frame
		{
			frame:          []float32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expectedOutput: false,
		},
		// Test case 3: Low energy frame
		{
			frame:          []float32{0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01, 0.01},
			expectedOutput: false,
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		result := VAD(tc.frame)

		// Check if the output matches the expected result
		if result != tc.expectedOutput {
			t.Errorf("VAD(%v) returned %v, expected %v", tc.frame, result, tc.expectedOutput)
		}
	}
}
