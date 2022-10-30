package netrc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadNetrc(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		scenario string
		path     string
		machines Machines
		err      error
	}{
		{
			scenario: "read simple netrc",
			path:     "testdata/example.netrc",
			machines: machines{
				"proxy.example.com": Details{
					Login:    "username",
					Password: "password",
				},
			},
			err: nil,
		},
		{
			scenario: "Duplicate login for one machine",
			path:     "testdata/duplicate_login.netrc",
			machines: nil,
			err:      ErrInvalid,
		},
		{
			scenario: "Duplicate password for one machine",
			path:     "testdata/duplicate_password.netrc",
			machines: nil,
			err:      ErrInvalid,
		},
	} {
		tc := tc
		t.Run(tc.scenario, func(t *testing.T) {
			t.Parallel()

			m, err := NewMachines(tc.path)
			assert.ErrorIs(t, err, tc.err, "Must match the expected error")
			assert.EqualValues(t, tc.machines, m, "Must match the expected machines")
		})
	}
}

func TestNewMachineFromEnv(t *testing.T) {
	t.Setenv(EnvironmentName, "testdata/example.netrc")

	m, err := NewMachinesFromEnvironment()
	assert.NoError(t, err, "Must not error reading environment")
	assert.NotNil(t, m, "Must be valid value")
}
