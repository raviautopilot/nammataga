package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"e2e-template/pkg/logger"
)

// HttpError defines the interface for HTTP errors returned by the client.
type HttpError interface {
	error
	StatusCode() int
	ResponseBody() string
	UnderlyingError() error
}

type httpErrorImpl struct {
	statusCode int
	respBody   string
	err        error
}

func (e *httpErrorImpl) Error() string {
	if e.err != nil {
		return fmt.Sprintf("HTTP Error: status=%d, err=%s", e.statusCode, e.err.Error())
	}
	return fmt.Sprintf("HTTP Error: status=%d", e.statusCode)
}

func (e *httpErrorImpl) StatusCode() int {
	return e.statusCode
}

func (e *httpErrorImpl) ResponseBody() string {
	return e.respBody
}

func (e *httpErrorImpl) UnderlyingError() error {
	return e.err
}

// Client represents the custom HTTP client.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	LogDir     string
	mu         sync.Mutex
	LastError  error
}

// NewClient initializes a new client with the given base URL and request logging directory.
func NewClient(baseURL string, timeout time.Duration, logDir string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		LogDir: logDir,
	}
}

// SendHttpRequest executes the HTTP request, validates input/output pointers, performs auth, and logs details.
func (c *Client) SendHttpRequest(method string, path string, headers map[string]string, reqBodyPtr interface{}, respBodyPtr interface{}, auth Authenticator) (httpErr HttpError) {
	defer func() {
		if httpErr != nil {
			c.LastError = httpErr
		}
	}()
	// 1. Validate pointers
	if reqBodyPtr != nil {
		val := reflect.ValueOf(reqBodyPtr)
		if val.Kind() != reflect.Ptr {
			return &httpErrorImpl{statusCode: 0, err: errors.New("request body must be a pointer to a struct/value")}
		}
		if val.IsNil() {
			return &httpErrorImpl{statusCode: 0, err: errors.New("request body pointer must not be nil")}
		}
	}

	if respBodyPtr != nil {
		val := reflect.ValueOf(respBodyPtr)
		if val.Kind() != reflect.Ptr {
			return &httpErrorImpl{statusCode: 0, err: errors.New("response body must be a pointer to a struct/value")}
		}
		if val.IsNil() {
			return &httpErrorImpl{statusCode: 0, err: errors.New("response body pointer must not be nil")}
		}
	}

	// 2. Configure transport for mTLS if ClientCertAuth is provided
	if auth != nil {
		if certAuth, ok := auth.(*ClientCertAuth); ok {
			c.mu.Lock()
			if c.HTTPClient.Transport == nil {
				c.HTTPClient.Transport = &http.Transport{
					TLSClientConfig: &tls.Config{
						Certificates: []tls.Certificate{certAuth.Certificate},
					},
				}
			} else {
				if transport, ok := c.HTTPClient.Transport.(*http.Transport); ok {
					if transport.TLSClientConfig == nil {
						transport.TLSClientConfig = &tls.Config{
							Certificates: []tls.Certificate{certAuth.Certificate},
						}
					} else {
						transport.TLSClientConfig.Certificates = []tls.Certificate{certAuth.Certificate}
					}
				}
			}
			c.mu.Unlock()
		}
	}

	// 3. Marshal Request Body
	var bodyReader io.Reader
	var requestRawBody []byte
	if reqBodyPtr != nil {
		var err error
		requestRawBody, err = json.Marshal(reqBodyPtr)
		if err != nil {
			return &httpErrorImpl{statusCode: 0, err: fmt.Errorf("failed to marshal request body: %w", err)}
		}
		bodyReader = bytes.NewReader(requestRawBody)
	}

	// 4. Create request
	var fullURL string
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		fullURL = path
	} else {
		fullURL = strings.TrimRight(c.BaseURL, "/") + "/" + strings.TrimLeft(path, "/")
	}
	httpReq, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return &httpErrorImpl{statusCode: 0, err: fmt.Errorf("failed to create http request: %w", err)}
	}

	// 5. Apply Headers
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	// 6. Apply Authentication
	if auth != nil {
		if err := auth.Apply(httpReq); err != nil {
			return &httpErrorImpl{statusCode: 0, err: fmt.Errorf("failed to apply authentication: %w", err)}
		}
	}

	// 7. Execute Request
	logger.Debug("Sending API request: %s %s", method, fullURL)
	startTime := time.Now()
	httpResp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		c.logExchange(startTime, httpReq, requestRawBody, nil, nil, err)
		return &httpErrorImpl{statusCode: 0, err: fmt.Errorf("network execution failed: %w", err)}
	}
	defer httpResp.Body.Close()

	// 8. Read Response Body
	responseRawBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		c.logExchange(startTime, httpReq, requestRawBody, httpResp, nil, err)
		return &httpErrorImpl{statusCode: httpResp.StatusCode, err: fmt.Errorf("failed to read response body: %w", err)}
	}

	// Log the API exchange
	c.logExchange(startTime, httpReq, requestRawBody, httpResp, responseRawBody, nil)

	// 9. Unmarshal Response Body if successful and pointer provided
	if httpResp.StatusCode >= 200 && httpResp.StatusCode < 300 {
		if respBodyPtr != nil && len(responseRawBody) > 0 {
			if err := json.Unmarshal(responseRawBody, respBodyPtr); err != nil {
				return &httpErrorImpl{
					statusCode: httpResp.StatusCode,
					respBody:   string(responseRawBody),
					err:        fmt.Errorf("failed to unmarshal response body: %w", err),
				}
			}
		}
		return nil
	}

	// 10. Return HTTP error for non-2xx status
	return &httpErrorImpl{
		statusCode: httpResp.StatusCode,
		respBody:   string(responseRawBody),
		err:        fmt.Errorf("request failed with status %d", httpResp.StatusCode),
	}
}

// logExchange logs the request/response details to disk.
func (c *Client) logExchange(startTime time.Time, req *http.Request, reqBody []byte, resp *http.Response, respBody []byte, err error) {
	now := time.Now()
	dateDir := now.Format("2006-01-02")
	timePrefix := now.Format("15-04-05")

	// Append suffix to avoid file collision during parallel runs
	rand.Seed(time.Now().UnixNano())
	suffix := fmt.Sprintf("%06d", rand.Intn(1000000))
	baseDir := filepath.Join(c.LogDir, dateDir)

	if err := os.MkdirAll(baseDir, 0777); err != nil {
		logger.Error("Failed to create log directory: %s", err)
		return
	}

	// Prepare Request JSON Log
	reqHeadersMap := make(map[string][]string)
	for k, v := range req.Header {
		reqHeadersMap[k] = v
	}

	var reqBodyJSON interface{}
	if len(reqBody) > 0 {
		_ = json.Unmarshal(reqBody, &reqBodyJSON)
	}

	reqLog := map[string]interface{}{
		"timestamp": startTime.Format(time.RFC3339Nano),
		"method":    req.Method,
		"url":       req.URL.String(),
		"headers":   reqHeadersMap,
		"body":      reqBodyJSON,
	}

	reqFilePath := filepath.Join(baseDir, fmt.Sprintf("%s-%s-request.json", timePrefix, suffix))
	reqFile, fileErr := os.Create(reqFilePath)
	if fileErr == nil {
		encoder := json.NewEncoder(reqFile)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(reqLog)
		reqFile.Close()
	}

	// Prepare Response JSON Log
	var respLog map[string]interface{}
	if resp != nil {
		respHeadersMap := make(map[string][]string)
		for k, v := range resp.Header {
			respHeadersMap[k] = v
		}

		var respBodyJSON interface{}
		if len(respBody) > 0 {
			_ = json.Unmarshal(respBody, &respBodyJSON)
		}

		respLog = map[string]interface{}{
			"timestamp":   now.Format(time.RFC3339Nano),
			"status_code": resp.StatusCode,
			"headers":     respHeadersMap,
			"body":        respBodyJSON,
			"latency_ms":  now.Sub(startTime).Milliseconds(),
		}
	} else {
		errMsg := "No response (network error)"
		if err != nil {
			errMsg = err.Error()
		}
		respLog = map[string]interface{}{
			"timestamp": now.Format(time.RFC3339Nano),
			"error":     errMsg,
		}
	}

	respFilePath := filepath.Join(baseDir, fmt.Sprintf("%s-%s-response.json", timePrefix, suffix))
	respFile, fileErr := os.Create(respFilePath)
	if fileErr == nil {
		encoder := json.NewEncoder(respFile)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(respLog)
		respFile.Close()
	}
}
