package logfetch

import (
	"bytes"
	"fmt"

	"github.com/greensnark/go-sequell/httpfetch"
	"github.com/greensnark/go-sequell/sources"
)

type FetchErrors []error

func (f FetchErrors) Error() string {
	buf := bytes.Buffer{}
	buf.WriteString("[")
	for i, e := range f {
		if i > 0 {
			buf.WriteString("; ")
			buf.WriteString(e.Error())
		}
	}
	buf.WriteString("]")
	return buf.String()
}

func sourceFetchRequests(incremental bool, src []*sources.XlogSrc) []*httpfetch.FetchRequest {
	res := make([]*httpfetch.FetchRequest, 0, len(src))
	for _, s := range src {
		if s.Local() {
			continue
		}
		if incremental && !s.Live && s.TargetExists() {
			continue
		}
		res = append(res, &httpfetch.FetchRequest{
			Url:      s.Url,
			Filename: s.TargetPath,
		})
	}
	return res
}

func Download(src *sources.Servers, incremental bool) error {
	req := sourceFetchRequests(incremental, src.XlogSources())
	nDownloads := len(req)
	fmt.Printf("Downloading %d files\n", nDownloads)
	resChan := httpfetch.New().ParallelFetch(req)

	errors := FetchErrors{}
	for fetchResult := range resChan {
		if fetchResult.Err != nil {
			errors = append(errors, fetchResult.Err)
		}
		ShowFetchResult(fetchResult)
	}
	if len(errors) == 0 {
		return nil
	} else {
		return errors
	}
}

func ShowFetchResult(res *httpfetch.FetchResult) {
	if res.Err != nil {
		fmt.Printf("ERR %s (%s)\n", res.Req, res.Err)
	} else {
		fmt.Printf("ok %s [%d]\n", res.Req, res.DownloadSize)
	}
}
