// Package bankilytest provides an in-process fake of the Bankily B-PAY API for
// controlled tests (timeouts, 5xx, malformed bodies, TS/TF/TA, token
// expiry). It is never wired into the server.
package bankilytest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Op is one payment the fake has recorded.
type Op struct {
	Status        string // TS | TF | TA
	TransactionID string
	Amount        string
	Phone         string
}

// Server is a fake Bankily. Zero-value hooks give the documented happy path:
// /payment returns errorCode 0 and records the op as DefaultStatus;
// /checkTransaction returns the recorded op, or errorCode 1 when unknown.
type Server struct {
	*httptest.Server

	mu  sync.Mutex
	ops map[string]*Op

	// defaultStatus is recorded for an accepted payment (default TS).
	defaultStatus string
	// TokenTTLSeconds / RefreshTTLSeconds are returned by /authentification.
	TokenTTLSeconds   int
	RefreshTTLSeconds int

	// Hooks (set via SetPaymentHook/SetCheckHook): return true when they
	// fully handled the response.
	paymentHook func(w http.ResponseWriter, body map[string]string) bool
	checkHook   func(w http.ResponseWriter, operationID string) bool

	validTokens  map[string]bool
	validRefresh map[string]bool
	tokenSeq     int

	PasswordGrants atomic.Int64
	RefreshGrants  atomic.Int64
	PaymentCalls   atomic.Int64
	CheckCalls     atomic.Int64
}

// New starts a fake server.
func New() *Server {
	s := &Server{
		ops:               map[string]*Op{},
		defaultStatus:     "TS",
		TokenTTLSeconds:   300,
		RefreshTTLSeconds: 1800,
		validTokens:       map[string]bool{},
		validRefresh:      map[string]bool{},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/authentification", s.auth)
	mux.HandleFunc("/payment", s.payment)
	mux.HandleFunc("/checkTransaction", s.check)
	s.Server = httptest.NewServer(mux)
	return s
}

// SetDefaultStatus sets the status recorded for newly accepted payments.
func (s *Server) SetDefaultStatus(status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.defaultStatus = status
}

// SetPaymentHook replaces the /payment hook (nil restores the default).
func (s *Server) SetPaymentHook(h func(w http.ResponseWriter, body map[string]string) bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.paymentHook = h
}

// SetCheckHook replaces the /checkTransaction hook (nil restores the default).
func (s *Server) SetCheckHook(h func(w http.ResponseWriter, operationID string) bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkHook = h
}

// SetOp forces the recorded state of an operation.
func (s *Server) SetOp(operationID, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	op, ok := s.ops[operationID]
	if !ok {
		op = &Op{TransactionID: "BK-" + operationID}
		s.ops[operationID] = op
	}
	op.Status = status
}

// GetOp returns a copy of the recorded operation.
func (s *Server) GetOp(operationID string) (Op, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	op, ok := s.ops[operationID]
	if !ok {
		return Op{}, false
	}
	return *op, true
}

// ExpireAccessTokens invalidates every issued access token (refresh tokens
// stay valid), simulating Bankily-side expiry.
func (s *Server) ExpireAccessTokens() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.validTokens = map[string]bool{}
}

// ExpireAllTokens invalidates access and refresh tokens.
func (s *Server) ExpireAllTokens() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.validTokens = map[string]bool{}
	s.validRefresh = map[string]bool{}
}

func (s *Server) issue(w http.ResponseWriter) {
	s.tokenSeq++
	at := fmt.Sprintf("at-%d", s.tokenSeq)
	rt := fmt.Sprintf("rt-%d", s.tokenSeq)
	s.validTokens[at] = true
	s.validRefresh[rt] = true
	writeJSON(w, 200, map[string]string{
		"access_token":       at,
		"expires_in":         fmt.Sprint(s.TokenTTLSeconds),
		"refresh_token":      rt,
		"refresh_expires_in": fmt.Sprint(s.RefreshTTLSeconds),
	})
}

func (s *Server) auth(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	s.mu.Lock()
	defer s.mu.Unlock()
	switch r.PostForm.Get("grant_type") {
	case "password":
		s.PasswordGrants.Add(1)
		if r.PostForm.Get("username") == "" || r.PostForm.Get("password") == "" {
			writeJSON(w, 401, map[string]string{"error": "invalid_grant"})
			return
		}
		s.issue(w)
	case "refresh_token":
		s.RefreshGrants.Add(1)
		if !s.validRefresh[r.PostForm.Get("refresh_token")] {
			writeJSON(w, 400, map[string]string{"error": "invalid_grant"})
			return
		}
		s.issue(w)
	default:
		writeJSON(w, 400, map[string]string{"error": "unsupported_grant_type"})
	}
}

func (s *Server) authorized(r *http.Request) bool {
	tok := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.validTokens[tok]
}

func (s *Server) payment(w http.ResponseWriter, r *http.Request) {
	s.PaymentCalls.Add(1)
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, 400, map[string]string{"errorCode": "1", "errorMessage": "bad json"})
		return
	}
	if !s.authorized(r) {
		writeJSON(w, 200, map[string]string{"errorCode": "2", "errorMessage": "invalid token"})
		return
	}
	s.mu.Lock()
	hook := s.paymentHook
	s.mu.Unlock()
	if hook != nil && hook(w, body) {
		return
	}
	op := body["operationId"]
	if op == "" {
		writeJSON(w, 200, map[string]string{"errorCode": "4", "errorMessage": "Operation ID required"})
		return
	}
	s.mu.Lock()
	if _, exists := s.ops[op]; exists {
		s.mu.Unlock()
		writeJSON(w, 200, map[string]string{"errorCode": "1", "errorMessage": "transaction existe deja"})
		return
	}
	rec := &Op{Status: s.defaultStatus, TransactionID: "BK-" + op, Amount: body["amount"], Phone: body["clientPhone"]}
	s.ops[op] = rec
	s.mu.Unlock()
	if rec.Status == "TF" {
		writeJSON(w, 200, map[string]string{"errorCode": "1", "errorMessage": "paiement refuse", "transactionId": rec.TransactionID})
		return
	}
	writeJSON(w, 200, map[string]string{"errorCode": "0", "errorMessage": "", "transactionId": rec.TransactionID})
}

func (s *Server) check(w http.ResponseWriter, r *http.Request) {
	s.CheckCalls.Add(1)
	var body map[string]string
	_ = json.NewDecoder(r.Body).Decode(&body)
	if !s.authorized(r) {
		writeJSON(w, 401, map[string]string{"error": "invalid token"})
		return
	}
	op := body["operationId"]
	s.mu.Lock()
	chook := s.checkHook
	s.mu.Unlock()
	if chook != nil && chook(w, op) {
		return
	}
	s.mu.Lock()
	rec, ok := s.ops[op]
	var cp Op
	if ok {
		cp = *rec
	}
	s.mu.Unlock()
	if !ok {
		writeJSON(w, 200, map[string]string{"errorCode": "1", "errorMessage": "transaction introuvable"})
		return
	}
	writeJSON(w, 200, map[string]string{"errorCode": "0", "errorMessage": "", "transactionId": cp.TransactionID, "status": cp.Status})
}

// Sleep is a hook helper that delays longer than a client timeout.
func Sleep(d time.Duration) { time.Sleep(d) }

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
