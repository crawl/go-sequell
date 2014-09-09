package logfetch

import (
	"fmt"

	"github.com/greensnark/go-sequell/httpfetch"
	"github.com/greensnark/go-sequell/sources"
)

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
	fmt.Printf("Downloading %d files\n", len(req))
	res := httpfetch.New().ParallelFetch(req)
	for _, r := range res {
		if r.Err != nil {
			return r.Err
		}
	}
	return nil
}
