package action

import (
	"os"

	"github.com/greensnark/go-sequell/crawl/data"
	"github.com/greensnark/go-sequell/logfetch"
	"github.com/greensnark/go-sequell/sources"
)

const LogCache = "server-xlogs"

func DownloadLogs(incremental bool) error {
	src, err := sources.Sources(data.Sources(), LogCache)
	if err != nil {
		return err
	}
	err = os.MkdirAll(LogCache, os.ModePerm)
	if err != nil {
		return err
	}
	return logfetch.Download(src, incremental)
}
