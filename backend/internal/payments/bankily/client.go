// Package bankily is the server-side HTTP client for the Bankily B-PAY
// e-commerce API (specification "B PAY version v1.0"):
//
//	POST /authentification  (x-www-form-urlencoded, OAuth2 password / refresh_token grants)
//	POST /payment           (JSON, Bearer)  -> errorCode, errorMessage, transactionId
//	POST /checkTransaction  (JSON, Bearer)  -> errorCode, errorMessage, transactionId, status
//
// Security: credentials and tokens live only inside Client. No method of
// this package logs, and no returned error ever embeds a credential, token,
// passcode or raw response body.
//
// Outcome classification: any failure where Bankily MAY have processed a
// payment (network error, timeout, 5xx, unreadable body) is returned as an
// error wrapping ErrUncertain. Callers must resolve those through
// CheckTransaction and must never blindly resend the payment.
package bankily

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Payment errorCode values (specification, "Paiement commerçant").
const (
	PaymentCodeSuccess             = "0"
	PaymentCodeOtherError          = "1"
	PaymentCodeInvalidToken        = "2"
	PaymentCodeOperationIDRequired = "4"
	// PaymentCodeOperationIDExists is NOT in the v1.0 specification; Bankily
	// TEST returned it ("Operation ID already existe") for a reused
	// operationId. A later TS for that operationId may not be ours.
	PaymentCodeOperationIDExists = "5"
)

// checkTransaction errorCode values and transaction statuses (specification,
// "Check de la transaction").
const (
	CheckCodeSuccess = "0"
	CheckCodeError   = "1"

	StatusSuccess = "TS" // transaction succès
	StatusFailed  = "TF" // transaction échouée
	StatusPending = "TA" // transaction en suspens
)

// tokenSkew renews an access token slightly before Bankily expires it.
const tokenSkew = 30 * time.Second

// maxResponseBytes caps how much of a Bankily response is read.
const maxResponseBytes = 1 << 20

var (
	// ErrUncertain: the outcome of the call is unknown -- Bankily may or may
	// not have processed it.
	ErrUncertain = errors.New("bankily: outcome uncertain")
	// ErrAuth: Bankily refused or could not issue an access token. Returned
	// by Payment only when /payment was never accepted for processing, so
	// the attempt can safely be considered not executed.
	ErrAuth = errors.New("bankily: authentication failed")
)

// UncertainError describes why an outcome is unknown, without any secret.
type UncertainError struct {
	Op     string // "payment" | "checkTransaction"
	Reason string // "timeout" | "network" | "http_5xx" | "malformed_response" | ...
	Err    error
}

func (e *UncertainError) Error() string {
	return fmt.Sprintf("bankily %s: outcome uncertain (%s)", e.Op, e.Reason)
}

func (e *UncertainError) Unwrap() []error {
	if e.Err == nil {
		return []error{ErrUncertain}
	}
	return []error{ErrUncertain, e.Err}
}

// Config holds the server-side connection settings.
type Config struct {
	BaseURL  string
	Username string
	Password string
	ClientID string
	Timeout  time.Duration
}

// PaymentRequest is the /payment input. Passcode is single-use and must never
// be persisted or logged by callers.
type PaymentRequest struct {
	ClientPhone string `json:"clientPhone"`
	Passcode    string `json:"passcode"`
	OperationID string `json:"operationId"`
	Amount      string `json:"amount"`
	Language    string `json:"language"` // EN | FR | AR
}

// PaymentResult is the /payment output.
type PaymentResult struct {
	ErrorCode     string
	ErrorMessage  string
	TransactionID string
}

// CheckResult is the /checkTransaction output.
type CheckResult struct {
	ErrorCode     string
	ErrorMessage  string
	TransactionID string
	Status        string
}

type tokenState struct {
	access     string
	accessExp  time.Time
	refresh    string
	refreshExp time.Time
}

// Client is safe for concurrent use. The access token is cached in memory and
// renewed through the refresh token, falling back to the password grant.
type Client struct {
	cfg  Config
	http *http.Client
	now  func() time.Time

	mu  sync.Mutex
	tok tokenState
}

// NewClient builds a client. httpClient may be nil.
func NewClient(cfg Config, httpClient *http.Client) *Client {
	if httpClient == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		httpClient = &http.Client{Timeout: timeout}
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	return &Client{cfg: cfg, http: httpClient, now: time.Now}
}

// flexString accepts a JSON string or number (Bankily documents every field
// as String, but numeric encodings are tolerated defensively).
type flexString string

func (f *flexString) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		*f = ""
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*f = flexString(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	*f = flexString(n.String())
	return nil
}

type tokenResponse struct {
	AccessToken      flexString `json:"access_token"`
	ExpiresIn        flexString `json:"expires_in"`
	RefreshToken     flexString `json:"refresh_token"`
	RefreshExpiresIn flexString `json:"refresh_expires_in"`
}

type apiResponse struct {
	ErrorCode     *flexString `json:"errorCode"`
	ErrorMessage  flexString  `json:"errorMessage"`
	TransactionID flexString  `json:"transactionId"`
	Status        flexString  `json:"status"`
}

// accessToken returns a valid access token, renewing it when needed.
func (c *Client) accessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	if c.tok.access != "" && now.Before(c.tok.accessExp) {
		return c.tok.access, nil
	}
	if c.tok.refresh != "" && now.Before(c.tok.refreshExp) {
		if err := c.authenticate(ctx, url.Values{
			"grant_type":    {"refresh_token"},
			"refresh_token": {c.tok.refresh},
			"client_id":     {c.cfg.ClientID},
		}); err == nil {
			return c.tok.access, nil
		}
		// A rejected/expired refresh token falls back to a full login.
	}
	if err := c.authenticate(ctx, url.Values{
		"grant_type": {"password"},
		"username":   {c.cfg.Username},
		"password":   {c.cfg.Password},
		"client_id":  {c.cfg.ClientID},
	}); err != nil {
		c.tok = tokenState{}
		return "", err
	}
	return c.tok.access, nil
}

// authenticate performs one /authentification call. Caller holds c.mu.
func (c *Client) authenticate(ctx context.Context, form url.Values) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/authentification", strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("%w: build request", ErrAuth)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrAuth, transportReason(err))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return fmt.Errorf("%w: read response", ErrAuth)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: http %d", ErrAuth, resp.StatusCode)
	}
	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil || tr.AccessToken == "" {
		return fmt.Errorf("%w: malformed token response", ErrAuth)
	}
	now := c.now()
	c.tok = tokenState{
		access:     string(tr.AccessToken),
		accessExp:  now.Add(durationOrDefault(string(tr.ExpiresIn), 60*time.Second) - tokenSkew),
		refresh:    string(tr.RefreshToken),
		refreshExp: now.Add(durationOrDefault(string(tr.RefreshExpiresIn), 0) - tokenSkew),
	}
	return nil
}

// invalidate drops a token Bankily rejected, but only if no other goroutine
// has already replaced it.
func (c *Client) invalidate(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.tok.access == token {
		c.tok.access = ""
		c.tok.accessExp = time.Time{}
	}
}

func durationOrDefault(secs string, def time.Duration) time.Duration {
	n, err := strconv.ParseFloat(strings.TrimSpace(secs), 64)
	if err != nil || n <= 0 {
		return def
	}
	return time.Duration(n * float64(time.Second))
}

// Payment calls /payment. A rejected token (HTTP 401 or errorCode 2) is
// renewed once and the SAME request (same operationId) retried -- Bankily
// rejected it before processing, so the retry cannot double-charge.
func (c *Client) Payment(ctx context.Context, p PaymentRequest) (*PaymentResult, error) {
	for attempt := 0; attempt < 2; attempt++ {
		token, err := c.accessToken(ctx)
		if err != nil {
			return nil, err
		}
		status, r, err := c.postJSON(ctx, "payment", "/payment", token, p)
		if err != nil {
			return nil, err
		}
		if status == http.StatusUnauthorized || (r != nil && r.ErrorCode != nil && string(*r.ErrorCode) == PaymentCodeInvalidToken) {
			c.invalidate(token)
			if attempt == 0 {
				continue
			}
			return nil, fmt.Errorf("%w: token rejected after renewal", ErrAuth)
		}
		if r == nil || r.ErrorCode == nil {
			return nil, &UncertainError{Op: "payment", Reason: "malformed_response"}
		}
		return &PaymentResult{
			ErrorCode:     strings.TrimSpace(string(*r.ErrorCode)),
			ErrorMessage:  string(r.ErrorMessage),
			TransactionID: strings.TrimSpace(string(r.TransactionID)),
		}, nil
	}
	return nil, fmt.Errorf("%w: token rejected", ErrAuth)
}

// CheckTransaction calls /checkTransaction. It is read-only on Bankily's
// side, so retrying it is always safe.
func (c *Client) CheckTransaction(ctx context.Context, operationID string) (*CheckResult, error) {
	for attempt := 0; attempt < 2; attempt++ {
		token, err := c.accessToken(ctx)
		if err != nil {
			return nil, &UncertainError{Op: "checkTransaction", Reason: "auth", Err: err}
		}
		status, r, err := c.postJSON(ctx, "checkTransaction", "/checkTransaction", token, map[string]string{"operationId": operationID})
		if err != nil {
			return nil, err
		}
		if status == http.StatusUnauthorized {
			c.invalidate(token)
			continue
		}
		if r == nil || r.ErrorCode == nil {
			return nil, &UncertainError{Op: "checkTransaction", Reason: "malformed_response"}
		}
		return &CheckResult{
			ErrorCode:     strings.TrimSpace(string(*r.ErrorCode)),
			ErrorMessage:  string(r.ErrorMessage),
			TransactionID: strings.TrimSpace(string(r.TransactionID)),
			Status:        strings.ToUpper(strings.TrimSpace(string(r.Status))),
		}, nil
	}
	return nil, &UncertainError{Op: "checkTransaction", Reason: "token_rejected"}
}

// postJSON returns (httpStatus, parsedBody, err). err is always an
// *UncertainError: transport failure, 5xx, or an unreadable non-401 body.
// A 401 is returned as status with a nil body for the caller to handle.
func (c *Client) postJSON(ctx context.Context, op, path, token string, payload any) (int, *apiResponse, error) {
	buf, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, &UncertainError{Op: op, Reason: "encode"}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+path, bytes.NewReader(buf))
	if err != nil {
		return 0, nil, &UncertainError{Op: op, Reason: "build_request"}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, &UncertainError{Op: op, Reason: transportReason(err), Err: ctx.Err()}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return 0, nil, &UncertainError{Op: op, Reason: "read_response"}
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return resp.StatusCode, nil, nil
	}
	if resp.StatusCode >= 500 {
		return resp.StatusCode, nil, &UncertainError{Op: op, Reason: fmt.Sprintf("http_%d", resp.StatusCode)}
	}
	var r apiResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return resp.StatusCode, nil, &UncertainError{Op: op, Reason: "malformed_response"}
	}
	return resp.StatusCode, &r, nil
}

func transportReason(err error) string {
	var ne interface{ Timeout() bool }
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &ne) && ne.Timeout()) {
		return "timeout"
	}
	return "network"
}
