package loader

import (
	"time"

	"github.com/crawl/go-sequell/crawl/ctime"
	"github.com/crawl/go-sequell/crawl/version"
	"github.com/crawl/go-sequell/xlog"
	"github.com/pkg/errors"
)

// A LogNormalizer cleans up an xlog record in some way
type LogNormalizer interface {
	NormalizeLog(log xlog.Xlog) error
}

// ReaderNormalizedLog adds log reader metadata to x and normalizes x using
// normalizer.
func ReaderNormalizedLog(reader *Reader, normalizer LogNormalizer, x xlog.Xlog) error {
	x["file"] = reader.Filename
	x["table"] = reader.Table
	x["base_table"] = reader.Type.BaseTable()
	x["src"] = reader.Server.Name
	x["offset"] = x[":offset"]
	delete(x, ":offset")

	if version.IsVersionLike(x["explbr"]) {
		delete(x, "explbr")
	}

	if err := normalizer.NormalizeLog(x); err != nil {
		return errors.Wrapf(err, "NormalizeLog(%#v)", x)
	}

	var err error
	normTime := func(field string) {
		if err != nil {
			return
		}

		timeStr, ok := x[field]
		if !ok {
			return
		}

		var logTime time.Time
		if logTime, err = reader.Server.ParseLogTime(timeStr); err != nil {
			err = errors.Wrapf(err, "normTime(%#v/%#v) in %#v", timeStr, field, x)
			return
		}

		x[field] = logTime.Format(ctime.LayoutDBTime)
	}

	normTime("start")
	normTime("end")
	normTime("time")

	if err != nil {
		return errors.Wrapf(err, "NormalizeLog(%#v)", x)
	}
	return nil
}
