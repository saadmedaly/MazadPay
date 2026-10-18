package bankily_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/mazadpay/backend/internal/payments/bankily"
	"github.com/mazadpay/backend/internal/payments/bankily/bankilytest"
)

func newClient(fake *bankilytest.Server, timeout time.Duration) *bankily.Client {
	return bankily.NewClient(bankily.Config{
		BaseURL: fake.URL, Username: "u", Password: "p", ClientID: "ebankily", Timeout: timeout,
	}, nil)
}

func pay(op string) bankily.PaymentRequest {
	return bankily.PaymentRequest{ClientPhone: "48110011", Passcode: "0000", OperationID: op, Amount: "120", Language: "FR"}
}

func TestPayment_Success_TokenCached(t *testing.T) {
	fake := bankilytest.New()
	defer fake.Close()
	c := newClient(fake, 5*time.Second)

	for _, op := range []string{"OP1", "OP2"} {
		res, err := c.Payment(context.Background(), pay(op))
		if err != nil || res.ErrorCode != bankily.PaymentCodeSuccess || res.TransactionID == "" {
			t.Fatalf("payment %s: res=%+v err=%v", op, res, err)
		}
	}
	if got := fake.PasswordGrants.Load(); got != 1 {
		t.Fatalf("expected the access token to be cached (1 login), got %d", got)
	}
	chk, err := c.CheckTransaction(context.Background(), "OP1")
	if err != nil || chk.Status != bankily.StatusSuccess {
		t.Fatalf("check: %+v %v", chk, err)
	}
}

func TestPayment_ExpiredAccessToken_UsesRefreshAndRetriesSameOperation(t *testing.T) {
	fake := bankilytest.New()
	defer fake.Close()
	c := newClient(fake, 5*time.Second)
	if _, err := c.Payment(context.Background(), pay("A1")); err != nil {
		t.Fatal(err)
	}
	fake.ExpireAccessTokens() // Bankily-side expiry: next /payment returns errorCode 2

	res, err := c.Payment(context.Background(), pay("A2"))
	if err != nil || res.ErrorCode != "0" {
		t.Fatalf("expected success after token renewal, got %+v %v", res, err)
	}
	if fake.PaymentCalls.Load() != 3 {
		t.Fatalf("expected exactly one retry of the rejected request, payment calls=%d", fake.PaymentCalls.Load())
	}
	if _, ok := fake.GetOp("A2"); !ok {
		t.Fatal("retried payment must keep the same operationId")
	}
}

func TestToken_ClientSideExpiry_UsesRefreshGrant(t *testing.T) {
	fake := bankilytest.New()
	fake.TokenTTLSeconds = 31 // minus 30s skew -> expires after ~1s
	defer fake.Close()
	c := newClient(fake, 5*time.Second)
	if _, err := c.CheckTransaction(context.Background(), "X"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1200 * time.Millisecond)
	if _, err := c.CheckTransaction(context.Background(), "X"); err != nil {
		t.Fatal(err)
	}
	if fake.RefreshGrants.Load() != 1 || fake.PasswordGrants.Load() != 1 {
		t.Fatalf("expected 1 login + 1 refresh, got login=%d refresh=%d", fake.PasswordGrants.Load(), fake.RefreshGrants.Load())
	}
}

func TestToken_RefreshRejected_FallsBackToPasswordGrant(t *testing.T) {
	fake := bankilytest.New()
	fake.TokenTTLSeconds = 31
	defer fake.Close()
	c := newClient(fake, 5*time.Second)
	if _, err := c.CheckTransaction(context.Background(), "X"); err != nil {
		t.Fatal(err)
	}
	fake.ExpireAllTokens()
	time.Sleep(1200 * time.Millisecond)
	if _, err := c.CheckTransaction(context.Background(), "X"); err != nil {
		t.Fatal(err)
	}
	if fake.PasswordGrants.Load() != 2 {
		t.Fatalf("expected a fresh login after refresh rejection, logins=%d", fake.PasswordGrants.Load())
	}
}

func TestPayment_Timeout_IsUncertainNotFailure(t *testing.T) {
	fake := bankilytest.New()
	defer fake.Close()
	c := newClient(fake, 150*time.Millisecond)
	// Warm the token before the slow path.
	_, _ = c.CheckTransaction(context.Background(), "warm")
	fake.SetPaymentHook(func(w http.ResponseWriter, body map[string]string) bool {
		time.Sleep(400 * time.Millisecond)
		return false
	})

	_, err := c.Payment(context.Background(), pay("T1"))
	var ue *bankily.UncertainError
	if !errors.Is(err, bankily.ErrUncertain) || !errors.As(err, &ue) || ue.Reason != "timeout" {
		t.Fatalf("timeout must be classified uncertain/timeout, got %v", err)
	}
	if errors.Is(err, bankily.ErrAuth) {
		t.Fatal("timeout must never look like a definite rejection")
	}
}

func TestPayment_5xx_And_Malformed_AreUncertain(t *testing.T) {
	fake := bankilytest.New()
	defer fake.Close()
	c := newClient(fake, 5*time.Second)

	fake.SetPaymentHook(func(w http.ResponseWriter, _ map[string]string) bool {
		w.WriteHeader(http.StatusBadGateway)
		return true
	})
	if _, err := c.Payment(context.Background(), pay("E1")); !errors.Is(err, bankily.ErrUncertain) {
		t.Fatalf("5xx must be uncertain, got %v", err)
	}

	fake.SetPaymentHook(func(w http.ResponseWriter, _ map[string]string) bool {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html>gateway</html>"))
		return true
	})
	if _, err := c.Payment(context.Background(), pay("E2")); !errors.Is(err, bankily.ErrUncertain) {
		t.Fatalf("malformed body must be uncertain, got %v", err)
	}

	fake.SetPaymentHook(func(w http.ResponseWriter, _ map[string]string) bool {
		_, _ = w.Write([]byte(`{"errorMessage":"no code"}`))
		return true
	})
	if _, err := c.Payment(context.Background(), pay("E3")); !errors.Is(err, bankily.ErrUncertain) {
		t.Fatalf("missing errorCode must be uncertain, got %v", err)
	}
}

func TestPayment_NumericErrorCode_Tolerated(t *testing.T) {
	fake := bankilytest.New()
	defer fake.Close()
	c := newClient(fake, 5*time.Second)
	fake.SetPaymentHook(func(w http.ResponseWriter, _ map[string]string) bool {
		_, _ = w.Write([]byte(`{"errorCode":0,"errorMessage":"","transactionId":12345}`))
		return true
	})
	res, err := c.Payment(context.Background(), pay("N1"))
	if err != nil || res.ErrorCode != "0" || res.TransactionID != "12345" {
		t.Fatalf("numeric fields: %+v %v", res, err)
	}
}

func TestPayment_BadCredentials_IsAuthErrorNotUncertain(t *testing.T) {
	fake := bankilytest.New()
	defer fake.Close()
	c := bankily.NewClient(bankily.Config{BaseURL: fake.URL, ClientID: "ebankily"}, nil)
	_, err := c.Payment(context.Background(), pay("B1"))
	if !errors.Is(err, bankily.ErrAuth) || errors.Is(err, bankily.ErrUncertain) {
		t.Fatalf("expected ErrAuth, got %v", err)
	}
	if fake.PaymentCalls.Load() != 0 {
		t.Fatal("/payment must not be called without a token")
	}
}

func TestErrors_NeverContainSecrets(t *testing.T) {
	fake := bankilytest.New()
	defer fake.Close()
	c := bankily.NewClient(bankily.Config{BaseURL: fake.URL, Username: "user-SECRET", Password: "pw-SECRET", ClientID: "ebankily"}, nil)
	fake.SetPaymentHook(func(w http.ResponseWriter, _ map[string]string) bool {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`{"access_token":"tok-SECRET"}`))
		return true
	})
	p := pay("S1")
	p.Passcode = "pass-SECRET"
	_, err := c.Payment(context.Background(), p)
	if err == nil {
		t.Fatal("expected error")
	}
	for _, s := range []string{"SECRET", "at-", "rt-"} {
		if containsFold(err.Error(), s) {
			t.Fatalf("error leaks sensitive material (%q): %v", s, err)
		}
	}
}

func containsFold(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && (stringIndex(s, sub) >= 0)
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
