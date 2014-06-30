package xlogtools

import (
	"github.com/greensnark/go-sequell/xlog"
	"testing"
)

var alphaLog xlog.XlogLine = xlog.XlogLine{
	"v":     "0.10-a0",
	"alpha": "y",
}

func TestNormalizeLogVersions(t *testing.T) {
	res := NormalizeLog(alphaLog.Clone())
	if res["cv"] != "0.10-a" {
		t.Errorf("Expected CV to be 0.10-a for %v, but was %s", alphaLog, res["cv"])
	}
	if res["v"] != "0.10.0-a0" {
		t.Errorf("Expected V to be 0.10.0-a0 for %v, but was %s", alphaLog, res["v"])
	}
}
