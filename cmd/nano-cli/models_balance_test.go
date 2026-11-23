package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/wishmatic/nano-mcp/internal/cli"
)

func TestModelsBalancePrintsTheAccountCredit(t *testing.T) {
	server, requests := captureServer(t, http.StatusOK,
		`{"usd_balance":"12.34","nano_balance":"5.6","nanoDepositAddress":"nano_abc"}`)

	out, _, err := runCommand(t, cli.Deps{Client: newClientAt(t, server.URL)}, "models", "balance")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	request := nextRequest(t, requests)

	if request.Method != http.MethodPost || request.Path != "/api/check-balance" {
		t.Errorf("request = %s %s, want POST /api/check-balance", request.Method, request.Path)
	}
	if key := request.Header.Get("x-api-key"); key != "test-key" {
		t.Errorf("x-api-key = %q, want test-key", key)
	}

	for _, want := range []string{"USD: $12.34", "Nano: 5.6", "Deposit address: nano_abc"} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout does not contain %q:\n%s", want, out)
		}
	}
}

func TestModelsBalanceOmitsAMissingDepositAddress(t *testing.T) {
	server, _ := captureServer(t, http.StatusOK, `{"usd_balance":"1","nano_balance":"2"}`)

	out, _, err := runCommand(t, cli.Deps{Client: newClientAt(t, server.URL)}, "models", "balance")
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	if strings.Contains(out, "Deposit address") {
		t.Errorf("stdout contains a missing deposit address:\n%s", out)
	}
}

func TestModelsBalanceReportsAnUpstreamFailure(t *testing.T) {
	server, _ := captureServer(t, http.StatusUnauthorized, `{"error":"bad key"}`)

	_, _, err := runCommand(t, cli.Deps{Client: newClientAt(t, server.URL)}, "models", "balance")
	if err == nil || !strings.Contains(err.Error(), "models balance") {
		t.Errorf("Execute() error = %v, want the upstream failure", err)
	}
}
