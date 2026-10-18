package database

import (
	"strings"
	"testing"

	"github.com/mazadpay/backend/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

const testRedisSecret = "s3cr3tP4ssw0rd"

func TestRedactRedisURL_RemovesCredentials(t *testing.T) {
	cases := map[string]string{
		"redis://red-abc:" + testRedisSecret + "@red-abc:6379":                          "redis://red-abc:6379",
		"rediss://default:" + testRedisSecret + "@x.upstash.io:6379/0":                  "rediss://x.upstash.io:6379/0",
		"redis://:" + testRedisSecret + "@localhost:6380/1?password=" + testRedisSecret: "redis://localhost:6380/1",
		"redis://localhost:6379/0":                                                      "redis://localhost:6379/0",
	}
	for in, want := range cases {
		got := RedactRedisURL(in)
		if got != want || strings.Contains(got, testRedisSecret) {
			t.Fatalf("RedactRedisURL: got %q want %q", got, want)
		}
	}
	for _, bad := range []string{"redis://u:" + testRedisSecret + "@host:notaport", "::" + testRedisSecret, ""} {
		if got := RedactRedisURL(bad); strings.Contains(got, testRedisSecret) {
			t.Fatalf("unparseable input leaked the secret: %q", got)
		}
	}
}

func TestNewRedis_ParseErrorNeverContainsSecret(t *testing.T) {
	cfg := &config.Config{}
	cfg.Redis.URL = "redis://default:" + testRedisSecret + "@host:notaport/0"
	_, err := NewRedis(cfg, zap.NewNop())
	if err == nil || strings.Contains(err.Error(), testRedisSecret) {
		t.Fatalf("parse error must not contain the password: %v", err)
	}
}

func TestNewRedis_ConnectFailureLogsNoSecret(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	cfg := &config.Config{}
	cfg.Redis.URL = "redis://default:" + testRedisSecret + "@127.0.0.1:1/0" // nothing listens on port 1
	_, err := NewRedis(cfg, zap.New(core))
	if err == nil {
		t.Skip("unexpected redis on 127.0.0.1:1")
	}
	if strings.Contains(err.Error(), testRedisSecret) {
		t.Fatalf("connect error leaked the password: %v", err)
	}
	for _, e := range logs.All() {
		for _, f := range e.Context {
			if strings.Contains(f.String, testRedisSecret) {
				t.Fatalf("log field leaked the password: %s=%s", f.Key, f.String)
			}
		}
	}
}
