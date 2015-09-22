package logfetch

import (
	"bytes"

	"github.com/crawl/go-sequell/httpfetch"
	"github.com/crawl/go-sequell/sources"
)

type XlogSourcePredicate interface {
	Match(src *sources.XlogSrc) bool
}

type FetchErrors []error

func (f FetchErrors) Error() string {
	buf := bytes.Buffer{}
	buf.WriteString("[")
	for i, e := range f {
		if i > 0 {
			buf.WriteString("; ")
		}
		buf.WriteString(e.Error())
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

		if err := s.MkdirTarget(); err != nil {
			panic(err)
		}
		res = append(res, &httpfetch.FetchRequest{
			URL:      s.URL,
			Filename: s.TargetPath,
		})
	}
	return res
}

type Fetcher struct {
	HTTPFetch *httpfetch.Fetcher
}

func New() *Fetcher {
	return &Fetcher{
		HTTPFetch: httpfetch.New(),
	}
}

func (f *Fetcher) DownloadAndWait(files []*sources.XlogSrc, incremental bool) {
	f.Download(files, incremental)
	f.HTTPFetch.Shutdown()
}

func (f *Fetcher) Download(files []*sources.XlogSrc, incremental bool) {
	req := sourceFetchRequests(incremental, files)
	f.HTTPFetch.QueueFetch(req)
}
