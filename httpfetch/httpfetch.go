package httpfetch

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const DefaultUserAgent = "Sequell httpfetch/1.0"

type HttpFetcher struct {
	HttpClient                   *http.Client
	Quiet                        bool
	ConnectTimeout               time.Duration
	ReadTimeout                  time.Duration
	UserAgent                    string
	MaxConcurrentRequestsPerHost int
	Logger                       *log.Logger
	logWriter                    Logger
}

func New() *HttpFetcher {
	writer := CreateLogger()
	return &HttpFetcher{
		HttpClient:                   DefaultHttpClient,
		ConnectTimeout:               DefaultConnectTimeout,
		ReadTimeout:                  DefaultReadTimeout,
		UserAgent:                    DefaultUserAgent,
		MaxConcurrentRequestsPerHost: 5,
		logWriter:                    writer,
		Logger:                       log.New(writer, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile),
	}
}

var DefaultConnectTimeout = 10 * time.Second
var DefaultReadTimeout = 20 * time.Second
var DefaultHttpTransport = http.Transport{
	Dial: dialer(DefaultConnectTimeout, DefaultReadTimeout),
	ResponseHeaderTimeout: DefaultConnectTimeout,
}
var DefaultHttpClient = &http.Client{
	Transport: &DefaultHttpTransport,
}

func (h *HttpFetcher) SetLogWriter(writer io.Writer) {
	h.logWriter.SetWriter(writer)
}

type unbufferedWriter struct {
	file *os.File
}

func (uw unbufferedWriter) Write(b []byte) (n int, err error) {
	n, err = uw.file.Write(b)
	if err != nil {
		_ = uw.file.Sync()
	}
	return
}

func (h *HttpFetcher) SetLogFile(file *os.File) {
	h.SetLogWriter(unbufferedWriter{file: file})
}

func (h *HttpFetcher) Logf(format string, rest ...interface{}) {
	h.Logger.Printf(format, rest...)
}

func (h *HttpFetcher) GetConcurrentRequestCount(count int) int {
	if count > h.MaxConcurrentRequestsPerHost {
		return h.MaxConcurrentRequestsPerHost
	}
	return count
}

type Headers map[string]string

type HttpError struct {
	StatusCode int
	Response   *http.Response
}

func (err *HttpError) Error() string {
	req := err.Response.Request
	return fmt.Sprint(req.Method, " ", req.URL, " failed: ", err.StatusCode)
}

type FetchRequest struct {
	Url      string
	Filename string

	// Don't try to resume downloads if this is set.
	FullDownload   bool
	RequestHeaders Headers
}

func (req *FetchRequest) Host() (string, error) {
	reqUrl, err := url.Parse(req.Url)
	if err != nil {
		return "", err
	}
	return reqUrl.Host, nil
}

func (req *FetchRequest) String() string {
	return fmt.Sprint(req.Url, " -> ", req.Filename)
}

type FetchResult struct {
	Req          *FetchRequest
	Err          error
	DownloadSize int64
}

func fetchError(req *FetchRequest, err error) *FetchResult {
	return &FetchResult{req, err, 0}
}

func (headers Headers) AddHeaders(h *http.Header) {
	for header, value := range headers {
		h.Add(header, value)
	}
}

func (headers Headers) Copy() Headers {
	res := make(Headers)
	for k, v := range headers {
		res[k] = v
	}
	return res
}

func HeadersWith(headers Headers, newHeader, newValue string) Headers {
	headerCopy := headers.Copy()
	headerCopy[newHeader] = newValue
	return headerCopy
}

func (h *HttpFetcher) FileGetResponse(url string, headers Headers) (*http.Response, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("User-Agent", h.UserAgent)
	if headers != nil {
		headers.AddHeaders(&request.Header)
	}
	h.Logf("FileGetResponse[%s]: pre-connect", url)
	resp, err := h.HttpClient.Do(request)
	h.Logf("FileGetResponse[%s]: connected: %v, %v", url, resp, err)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode >= 400 {
		return nil, &HttpError{resp.StatusCode, resp}
	}
	return resp, err
}

func (h *HttpFetcher) FetchFile(req *FetchRequest, complete chan<- *FetchResult) {
	h.Logf("FetchFile[%s] -> %s (full download: %v)", req.Url, req.Filename,
		req.FullDownload)
	if !req.FullDownload {
		finf, err := os.Stat(req.Filename)
		if err == nil && finf != nil && finf.Size() > 0 {
			h.Logf("FetchFile[%s]: resuming download for %s",
				req.Url, req.Filename)
			h.ResumeFileDownload(req, complete)
			return
		}
	}
	h.Logf("FetchFile[%s]: new file download %s", req.Url, req.Filename)
	h.NewFileDownload(req, complete)
}

func fileResumeHeaders(req *FetchRequest, file *os.File) (Headers, int64) {
	headers := req.RequestHeaders
	finf, err := file.Stat()
	resumePoint := int64(0)
	if err == nil && finf != nil {
		resumePoint = finf.Size()
		if headers == nil {
			headers = Headers{}
		} else {
			headers = headers.Copy()
		}
		headers["Range"] = fmt.Sprintf("bytes=%d-", resumePoint)
		headers["Accept-Encoding"] = ""
	}
	return headers, resumePoint
}

func (h *HttpFetcher) ResumeFileDownload(req *FetchRequest, complete chan<- *FetchResult) {
	h.Logf("ResumeFileDownload[%s] -> %s", req.Url, req.Filename)
	var err error
	handleError := func() {
		if err != nil && !h.Quiet {
			h.Logf("Download of %s failed: %s\n", req, err)
		}
		h.Logf("handleError[%s] -> %s, err: %v", req.Url, req.Filename, err)
		complete <- fetchError(req, err)
	}

	if !h.Quiet {
		h.Logf("ResumeFileDownload(%s)\n", req)
	}
	file, err := os.OpenFile(req.Filename,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		handleError()
		return
	}
	defer file.Close()

	headers, resumePoint := fileResumeHeaders(req, file)
	resp, err := h.FileGetResponse(req.Url, headers)

	var copied int64 = 0
	if err != nil {
		httpErr, _ := err.(*HttpError)
		if httpErr == nil || httpErr.StatusCode != http.StatusRequestedRangeNotSatisfiable {
			handleError()
			return
		}
		err = nil
	} else {
		defer resp.Body.Close()
		h.Logf("ResumeFileDownload[%s]: Copying bytes to %s from response",
			req.Url, req.Filename)

		copied, err = io.Copy(file, resp.Body)
	}
	if !h.Quiet {
		h.Logf("[DONE:%d] ResumeFileDownload (at %d) %s\n", copied, resumePoint, req)
	}
	h.Logf("ResumeFileDownload[%s] -> %s: completed, bytes copied: %v, err: %v",
		req.Url, req.Filename, copied, err)
	complete <- &FetchResult{req, err, copied}
}

func (h *HttpFetcher) NewFileDownload(req *FetchRequest, complete chan<- *FetchResult) {
	h.Logf("NewFileDownload[%s] -> %s", req.Url, req.Filename)
	if !h.Quiet {
		h.Logf("NewFileDownload ", req)
	}
	resp, err := h.FileGetResponse(req.Url, req.RequestHeaders)
	if err != nil {
		h.Logf("NewFileDownload[%s] -> %s: error: %v (http)", req.Url, req.Filename, err)
		complete <- fetchError(req, err)
		return
	}
	defer resp.Body.Close()

	file, err := os.Create(req.Filename)
	if err != nil {
		h.Logf("NewFileDownload[%s] -> %s: error: %v (fopen)", req.Url, req.Filename, err)
		complete <- fetchError(req, err)
		return
	}
	defer file.Close()

	h.Logf("NewFileDownload[%s] -> %s: copying bytes", req.Url, req.Filename)
	copied, err := io.Copy(file, resp.Body)
	if !h.Quiet {
		h.Logf("[DONE:%d] NewFileDownload %s\n", copied, req)
	}
	h.Logf("NewFileDownload[%s] -> %s: completed copy:%v, err:%v", req.Url, req.Filename, copied, err)
	complete <- &FetchResult{req, err, copied}
}

func groupFetchRequestsByHost(requests []*FetchRequest) map[string][]*FetchRequest {
	grouped := make(map[string][]*FetchRequest)
	for _, req := range requests {
		reqHost, _ := req.Host()
		grouped[reqHost] = append(grouped[reqHost], req)
	}
	return grouped
}

func (h *HttpFetcher) ParallelFetch(requests []*FetchRequest) <-chan *FetchResult {
	h.Logf("ParallelFetch: %d files", len(requests))

	completion := make(chan *FetchResult)
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(requests))
	go func() {
		waitGroup.Wait()
		close(completion)
	}()

	runHostRequests := func(host string, reqs []*FetchRequest) {
		reqChan := make(chan *FetchRequest, len(reqs))
		for _, req := range reqs {
			reqChan <- req
		}
		close(reqChan)
		nRoutines := h.GetConcurrentRequestCount(len(reqs))
		h.Logf("runHostRequests[%s]: spinning up %d fetch routines for %d URLs",
			host, nRoutines, len(reqs))
		for i := 0; i < nRoutines; i++ {
			go func() {
				for req := range reqChan {
					h.FetchFile(req, completion)
					waitGroup.Done()
				}
			}()
		}
	}
	for host, requests := range groupFetchRequestsByHost(requests) {
		h.Logf("runHostRequests: %s (%d requests)", host, len(requests))
		runHostRequests(host, requests)
	}
	return completion
}

func UrlFilename(url string) (string, error) {
	for {
		if len(url) == 0 {
			return "", errors.New(fmt.Sprintf("No filename for empty URL"))
		}

		slashIndex := strings.LastIndex(url, "/")
		if slashIndex == -1 {
			return "", errors.New(fmt.Sprint("Cannot determine URL filename from ", url))
		}

		filename := url[slashIndex+1:]
		if len(filename) == 0 {
			url = url[:len(url)-1]
			continue
		}
		return filename, nil
	}
}
