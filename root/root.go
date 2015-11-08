package root

import (
	"io/ioutil"
	"os"
	"path"
)

// A Root is the tree root for Sequell's configuration and working directories.
type Root string

// New creates a root defaulting to defroot, overridden by the values of any
// of the given envVars, where the first-non-empty var wins.
func New(defroot string, envVars ...string) Root {
	root := defroot
	for _, env := range envVars {
		if value := os.Getenv(env); value != "" {
			root = value
			break
		}
	}
	if root == "" {
		var err error
		if root, err = os.Getwd(); err != nil {
			panic(err)
		}
	}
	return Root(root)
}

// Root gets the root directory path
func (r Root) Root() string { return string(r) }

// Path converts filepath to a path under r.
func (r Root) Path(filepath string) string {
	return path.Join(string(r), filepath)
}

// Bytes reads the file at path in r as a []byte
func (r Root) Bytes(path string) ([]byte, error) {
	bytes, err := ioutil.ReadFile(r.Path(path))
	if err != nil {
		return nil, err
	}
	return bytes, err
}

// String reads the file at path in r as a string
func (r Root) String(path string) (string, error) {
	bytes, err := r.Bytes(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
