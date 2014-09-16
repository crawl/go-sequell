package httpfetch

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

const DefaultUserAgent = "Sequell httpfetch/1.0"

type Fetcher struct {
	HTTPClient                   *http.Client
	Quiet                        bool
	ConnectTimeout               time.Duration
	ReadTimeout                  time.Duration
	UserAgent                    string
	MaxConcurrentRequestsPerHost int
	Logger                       *log.Logger
	logWriter                    Logger

	// Queues for each host, monitored by the service goroutine.
	hostQueues       map[string]chan<- *FetchRequest
	hostWaitGroup    sync.WaitGroup
	enqueueWaitGroup sync.WaitGroup
}

// New returns a new Fetcher for parallel downloads. Fetcher
// methods are not threadsafe.
func New() *Fetcher {
	writer := CreateLogger()
	return &Fetcher{
		HTTPClient:                   DefaultHTTPClient,
		ConnectTimeout:               DefaultConnectTimeout,
		ReadTimeout:                  DefaultReadTimeout,
		UserAgent:                    DefaultUserAgent,
		MaxConcurrentRequestsPerHost: 5,
		logWriter:                    writer,
		Logger:                       log.New(writer, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile),
		hostQueues:                   map[string]chan<- *FetchRequest{},
	}
}

var DefaultConnectTimeout = 10 * time.Second
var DefaultReadTimeout = 20 * time.Second
var DefaultHTTPTransport = http.Transport{
	Dial: dialer(DefaultConnectTimeout, DefaultReadTimeout),
	ResponseHeaderTimeout: DefaultConnectTimeout,
}
var DefaultHTTPClient = &http.Client{
	Transport: &DefaultHTTPTransport,
}

func (h *Fetcher) SetLogWriter(writer io.Writer) {
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

func (h *Fetcher) SetLogFile(file *os.File) {
	h.SetLogWriter(unbufferedWriter{file: file})
}

func (h *Fetcher) Logf(format string, rest ...interface{}) {
	h.Logger.Printf(format, rest...)
}

func (h *Fetcher) GetConcurrentRequestCount(count int) int {
	if count > h.MaxConcurrentRequestsPerHost {
		return h.MaxConcurrentRequestsPerHost
	}
	return count
}

type Headers map[string]string

type HTTPError struct {
	StatusCode int
	Response   *http.Response
}

func (err *HTTPError) Error() string {
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

func (h *Fetcher) FileGetResponse(url string, headers Headers) (*http.Response, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("User-Agent", h.UserAgent)
	if headers != nil {
		headers.AddHeaders(&request.Header)
	}
	h.Logf("FileGetResponse[%s]: pre-connect", url)
	resp, err := h.HTTPClient.Do(request)
	h.Logf("FileGetResponse[%s]: connected: %v, %v", url, resp, err)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		resp.Body.Close()
		return nil, &HTTPError{resp.StatusCode, resp}
	}
	return resp, err
}

func (h *Fetcher) FetchFile(req *FetchRequest, complete chan<- *FetchResult) {
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

func (h *Fetcher) ResumeFileDownload(req *FetchRequest, complete chan<- *FetchResult) {
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
		httpErr, _ := err.(*HTTPError)
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

func (h *Fetcher) NewFileDownload(req *FetchRequest, complete chan<- *FetchResult) {
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

// QueueFetch enqueues the given download requests for asynchronous download.
func (h *Fetcher) QueueFetch(req []*FetchRequest) {
	log.Printf("QueueFetch: %d files\n", len(req))
	for host, reqs := range groupFetchRequestsByHost(req) {
		hostQueue := h.hostQueue(host)
		h.enqueueWaitGroup.Add(1)
		go h.enqueueRequests(hostQueue, reqs)
	}
}

// Shutdown gracefully shuts down the fetcher, cleaning up all
// background goroutines, and waiting for all outstanding downloads to
// end.
func (h *Fetcher) Shutdown() {
	h.enqueueWaitGroup.Wait()
	for host, queue := range h.hostQueues {
		close(queue)
		delete(h.hostQueues, host)
	}
	h.hostWaitGroup.Wait()
}

func (h *Fetcher) enqueueRequests(queue chan<- *FetchRequest, reqs []*FetchRequest) {
	for _, req := range reqs {
		queue <- req
	}
	h.enqueueWaitGroup.Done()
}

func (h *Fetcher) hostQueue(host string) chan<- *FetchRequest {
	queue := h.hostQueues[host]
	if queue == nil {
		h.hostWaitGroup.Add(1)
		newQueue := make(chan *FetchRequest)
		go h.monitorHostQueue(host, newQueue)
		h.hostQueues[host] = newQueue
		queue = newQueue
	}
	return queue
}

func (h *Fetcher) monitorHostQueue(host string, incoming <-chan *FetchRequest) {
	slaveResult := make(chan *FetchResult)
	slaveQueue := make(chan *FetchRequest)

	nSlaves := h.MaxConcurrentRequestsPerHost
	slaveWaitGroup := sync.WaitGroup{}
	slaveWaitGroup.Add(nSlaves)
	// Slaves lead uncomplicated lives:
	for i := 0; i < nSlaves; i++ {
		go func() {
			for req := range slaveQueue {
				h.FetchFile(req, slaveResult)
			}
			slaveWaitGroup.Done()
		}()
	}
	// And a goroutine to close the slaveResult channel when
	// everyone's done.
	go func() {
		slaveWaitGroup.Wait()
		log.Printf("Cleaning up host monitor for %s\n", host)
		close(slaveResult)
	}()

	queue := []*FetchRequest{}
	inProgress := map[string]bool{}
	reqKey := func(req *FetchRequest) string {
		return req.Url + " | " + req.Filename
	}

	queueRequest := func(req *FetchRequest) {
		// Suppress duplicate fetch requests:
		key := reqKey(req)
		if inProgress[key] {
			log.Printf("%s: ignoring duplicate download %s\n", host, req.Url)
			return
		}
		inProgress[key] = true
		queue = append(queue, req)
	}

	applyResult := func(res *FetchResult) {
		delete(inProgress, reqKey(res.Req))
		if res.Err != nil {
			log.Printf("ERR %s (%s)\n", res.Req, res.Err)
		} else if res.DownloadSize > 0 {
			log.Printf("ok %s [%d]\n", res.Req, res.DownloadSize)
		}
	}

	firstItem := func() *FetchRequest {
		if len(queue) == 0 {
			return nil
		}
		return queue[0]
	}
	slaveQueueOrNil := func() chan<- *FetchRequest {
		if len(queue) == 0 {
			return nil
		}
		return slaveQueue
	}

	for incoming != nil || len(inProgress) > 0 {
		// Bi-modal select: if there are things in the queue, try to
		// feed them to the first slave who will listen. In all cases,
		// track incoming requests and slaves reporting results.
		select {
		case slaveQueueOrNil() <- firstItem():
			queue = queue[1:]
		case newRequest := <-incoming:
			if newRequest == nil {
				log.Printf("%s: Download queue shutting down\n", host)
				incoming = nil
				break
			}
			queueRequest(newRequest)
		case result := <-slaveResult:
			applyResult(result)
		}
	}

	// Exiting, clean up:
	close(slaveQueue)
	for res := range slaveResult {
		applyResult(res)
	}
	h.hostWaitGroup.Done()
}
