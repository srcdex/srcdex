// Package vector provides the vector maths shared by the embeddings
// engines and their consumers.
package vector

import (
	"math"
)

// L2Norm reports the Euclidean length of the vector; embedding
// models commonly return unit-normalised vectors.
func L2Norm(v []float32) float64 {
	var sum float64
	for _, x := range v {
		sum += float64(x) * float64(x)
	}
	return math.Sqrt(sum)
}
