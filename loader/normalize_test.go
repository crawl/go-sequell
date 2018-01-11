package loader

import (
	"fmt"
	"testing"

	"github.com/crawl/go-sequell/crawl/xlogtools"
	"github.com/crawl/go-sequell/sources"
	"github.com/crawl/go-sequell/xlog"
)

type normalizerFunc func(xlog.Xlog) error

func (n normalizerFunc) NormalizeLog(x xlog.Xlog) error {
	return n(x)
}

func TestReaderNormalizedLog(t *testing.T) {
	reader := &Reader{
		Reader: &xlog.Reader{
			Filename: "test.xlog",
		},
		XlogSrc: &sources.XlogSrc{
			Server: &sources.Server{
				Name: "test",
			},
			Type: xlogtools.Log,
		},
	}

	normalizeStub := normalizerFunc(func(x xlog.Xlog) error {
		x["normalized"] = "yes"
		return nil
	})

	for _, test := range []struct {
		name           string
		xlog           xlog.Xlog
		normalizedXlog xlog.Xlog
	}{
		{"offset", xlog.Xlog{":offset": "39"}, xlog.Xlog{"offset": "39"}},
		{"explbr-version-zap", xlog.Xlog{"explbr": "0.17"}, xlog.Xlog{"explbr": ""}},
		{"explbr-normal-passthrough", xlog.Xlog{"explbr": "direcows"}, xlog.Xlog{"explbr": "direcows"}},
	} {
		t.Run(fmt.Sprintf("TestReaderNormalizedLog(%#v)", test.name), func(t *testing.T) {
			if err := ReaderNormalizedLog(reader, normalizeStub, test.xlog); err != nil {
				t.Fatalf("err: %s", err)
			}

			for key, expectedValue := range test.normalizedXlog {
				if actualValue := test.xlog[key]; actualValue != expectedValue {
					t.Fatalf("normalized=%s, expected=%s (want %s=%s, got %s=%s)", test.xlog, test.normalizedXlog,
						key, expectedValue, key, actualValue)
				}
			}
		})
	}
}
