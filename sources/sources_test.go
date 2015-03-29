package sources

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestMakeTargetDir(t *testing.T) {
	dir, err := ioutil.TempDir("", "xlogtarget")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	for _, c := range []string{
		"cow", "yak/foo", "moo/path/zag.logfile",
	} {
		x := XlogSrc{TargetPath: filepath.Join(dir, c)}
		expectedDir := filepath.Dir(filepath.Join(dir, c))
		err := x.MkdirTarget()
		if err != nil {
			t.Errorf("&XlogSrc{TargetPath:%#v}.MakeTargetDir() failed for %#v: %s", x.TargetPath, c, err)
			continue
		}
		_, err = os.Stat(expectedDir)
		if err != nil {
			t.Errorf("&XlogSrc{TargetPath:%#v}.MakeTargetDir() failed: expected %s to exist, but it does not", x.TargetPath, expectedDir)
		}
	}
}
