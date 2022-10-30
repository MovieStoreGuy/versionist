package generic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcurrent(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		scenario string
		content  map[string]int
		fn       func(string, int) error
	}{
		{
			scenario: "Processing all values in list",
			content: map[string]int{
				"foo": 1,
				"bar": 2,
				"baz": 3,
			},
			fn: func(s string, i int) error {
				return nil
			},
		},
	} {
		t.Run(tc.scenario, func(t *testing.T) {
			assert.NoError(t, ParallelRangeMap(tc.content, tc.fn), "Must not error when processing map")
		})
	}
}
