package netrc

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
)

const (
	EnvironmentName = "NETRC"
)

var (
	ErrInvalid = errors.New("invalid netrc")
)

type (
	Details struct {
		Login    string
		Password string
	}

	Machines interface {
		GetMachineDetails(hostname string) (Details, bool)
	}

	machines map[string]Details
)

var (
	_ Machines = (*machines)(nil)
)

func NewMachinesFromEnvironment() (Machines, error) {
	if path, ok := os.LookupEnv(EnvironmentName); ok {
		return NewMachines(path)
	}
	dir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return NewMachines(path.Join(dir, ".netrc"))
}

func NewMachines(filepath string) (Machines, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	var (
		scanner = bufio.NewScanner(f)

		inMacro  = false
		host     string
		machines = make(machines)
		entry    = &Details{}
	)

	for scanner.Scan() {
		text := scanner.Text()
		if inMacro {
			if text == "" {
				inMacro = true
			}
			continue
		}

		fields := strings.Fields(text)
		for i := 0; i < len(fields)-1; i += 2 {
			switch fields[i] {
			case "machine":
				host = fields[i+1]
			case "login":
				if entry.Login != "" {
					return nil, fmt.Errorf("duplicate entries for login: %w", ErrInvalid)
				}
				entry.Login = fields[i+1]
			case "password":
				if entry.Password != "" {
					return nil, fmt.Errorf("duplicate entries for password: %w", ErrInvalid)
				}
				entry.Password = fields[i+1]
			case "macdef":
				inMacro = true
			}
			if host != "" && entry.Login != "" && entry.Password != "" {
				if _, exist := machines[host]; exist {
					return nil, fmt.Errorf("duplicate machine set: %w", ErrInvalid)
				}
				machines[host] = *entry
				host = ""
			}
		}
	}
	return machines, nil
}

func (m machines) GetMachineDetails(hostname string) (Details, bool) {
	v, ok := m[hostname]
	return v, ok
}
