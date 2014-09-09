package root

import (
	"io/ioutil"
	"os"
	"path"
)

type Root string

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

func (r Root) Root() string { return string(r) }
func (r Root) Path(filepath string) string {
	return path.Join(string(r), filepath)
}
func (r Root) Bytes(path string) ([]byte, error) {
	bytes, err := ioutil.ReadFile(r.Path(path))
	if err != nil {
		return nil, err
	}
	return bytes, err
}

func (r Root) String(path string) (string, error) {
	bytes, err := r.Bytes(path)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
