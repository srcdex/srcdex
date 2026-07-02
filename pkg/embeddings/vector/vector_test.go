package vector_test

import (
	"testing"

	"darvaza.org/core"

	"srcdex.dev/pkg/embeddings/vector"
)

// Compile-time verification that test case types implement the
// TestCase interface.
var _ core.TestCase = l2NormTestCase{}

// l2NormTestCase exercises L2Norm on a vector with a known length.
type l2NormTestCase struct {
	name string
	v    []float32
	want float64
}

func (tc l2NormTestCase) Name() string {
	return tc.name
}

func (tc l2NormTestCase) Test(t *testing.T) {
	t.Helper()

	core.AssertEqual(t, tc.want, vector.L2Norm(tc.v), "norm")
}

func newL2NormTestCase(name string, v []float32, want float64) l2NormTestCase {
	return l2NormTestCase{
		name: name,
		v:    v,
		want: want,
	}
}

func TestL2Norm(t *testing.T) {
	testCases := []l2NormTestCase{
		newL2NormTestCase("nil vector", nil, 0),
		newL2NormTestCase("zero vector", []float32{0, 0, 0}, 0),
		newL2NormTestCase("unit basis", []float32{0, 1, 0}, 1),
		newL2NormTestCase("3-4-5 triangle", []float32{3, 4}, 5),
	}

	core.RunTestCases(t, testCases)
}
