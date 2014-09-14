package logfetch

import (
	"bytes"

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
		res = append(res, &httpfetch.FetchRequest{
			Url:      s.Url,
			Filename: s.TargetPath,
		})
	}
	return res
}

type Fetcher struct {
	Servers   *sources.Servers
	HTTPFetch *httpfetch.Fetcher
}

func New(src *sources.Servers) *Fetcher {
	return &Fetcher{
		Servers:   src,
		HTTPFetch: httpfetch.New(),
	}
}

func (f *Fetcher) DownloadAndWait(incremental bool) {
	f.Download(incremental)
	f.HTTPFetch.Shutdown()
}

func (f *Fetcher) Download(incremental bool) {
	req := sourceFetchRequests(incremental, f.Servers.XlogSources())
	f.HTTPFetch.QueueFetch(req)
}
