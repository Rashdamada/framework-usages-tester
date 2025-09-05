// main.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

/*
MCP-style JSON-RPC 2.0 server over stdio with mock-banking domain.
Exposes:
- "ping"
- "resources/list"
- "resources/read"
- "tools/call" (list_accounts, get_balance, list_transactions, create_transfer)

This is dependency-free and keeps state in-memory.
*/

const jsonrpcVersion = "2.0"

// ---------- JSON-RPC Types ----------
type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      any         `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func okResult(id any, v interface{}) rpcResponse {
	return rpcResponse{JSONRPC: jsonrpcVersion, ID: id, Result: v}
}

func errResult(id any, code int, msg string, data any) rpcResponse {
	return rpcResponse{JSONRPC: jsonrpcVersion, ID: id, Error: &rpcError{Code: code, Message: msg, Data: data}}
}

// ---------- MCP-ish Types ----------
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type TextResourceContents struct {
	URI  string `json:"uri"`
	Text string `json:"text"`
}

// ---------- Banking Domain (Mock State) ----------
type Account struct {
	AccountID string `json:"accountId"`
	Type      string `json:"type"`
	Currency  string `json:"currency"`
	OwnerName string `json:"ownerName"`
}

type Balance struct {
	AccountID string  `json:"accountId"`
	Available float64 `json:"available"`
	Ledger    float64 `json:"ledger"`
	Currency  string  `json:"currency"`
}

type Transaction struct {
	TransactionID string  `json:"transactionId"`
	Date          string  `json:"date"` // RFC3339
	Description   string  `json:"description"`
	Amount        float64 `json:"amount"` // negative=debit, positive=credit
	Currency      string  `json:"currency"`
	Type          string  `json:"type"` // "debit"|"credit"
}

type TransferRequest struct {
	FromAccountID string  `json:"fromAccountId"`
	ToAccountID   string  `json:"toAccountId"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Description   string  `json:"description,omitempty"`
}

type TransferResponse struct {
	TransferID string `json:"transferId"`
	Status     string `json:"status"`    // "pending"|"completed"|"failed"
	Timestamp  string `json:"timestamp"` // RFC3339
}

// In-memory bank
type bankState struct {
	mu          sync.Mutex
	accounts    map[string]Account
	balances    map[string]Balance
	transactions map[string][]Transaction // by accountId
}

func newBank() *bankState {
	now := time.Now().UTC().Format(time.RFC3339)

	accts := map[string]Account{
		"CHK-001": {AccountID: "CHK-001", Type: "Checking", Currency: "USD", OwnerName: "Jane Doe"},
		"SAV-001": {AccountID: "SAV-001", Type: "Savings", Currency: "USD", OwnerName: "Jane Doe"},
	}
	bals := map[string]Balance{
		"CHK-001": {AccountID: "CHK-001", Available: 1250.75, Ledger: 1300.00, Currency: "USD"},
		"SAV-001": {AccountID: "SAV-001", Available: 5000.00, Ledger: 5000.00, Currency: "USD"},
	}
	txn := map[string][]Transaction{
		"CHK-001": {
			{TransactionID: "tx-1001", Date: now, Description: "Coffee shop", Amount: -4.50, Currency: "USD", Type: "debit"},
			{TransactionID: "tx-1002", Date: now, Description: "Payroll", Amount: 2500.00, Currency: "USD", Type: "credit"},
		},
		"SAV-001": {
			{TransactionID: "tx-2001", Date: now, Description: "Initial deposit", Amount: 5000.00, Currency: "USD", Type: "credit"},
		},
	}

	return &bankState{
		accounts:     accts,
		balances:     bals,
		transactions: txn,
	}
}

// ---------- MCP Handlers ----------
func handlePing(_ json.RawMessage) (any, *rpcError) {
	return map[string]string{"message": "pong"}, nil
}

func handleResourcesList(bank *bankState, _ json.RawMessage) (any, *rpcError) {
	// Expose top-level and per-account resources
	bank.mu.Lock()
	defer bank.mu.Unlock()

	var out []Resource
	out = append(out, Resource{
		URI:         "bank://accounts",
		Name:        "Accounts",
		Description: "List of accounts for the current user",
	})
	for id := range bank.accounts {
		out = append(out, Resource{
			URI:         fmt.Sprintf("bank://accounts/%s", id),
			Name:        fmt.Sprintf("Account %s", id),
			Description: "Account details",
		})
		out = append(out, Resource{
			URI:         fmt.Sprintf("bank://accounts/%s/balance", id),
			Name:        fmt.Sprintf("Balance %s", id),
			Description: "Current account balance",
		})
		out = append(out, Resource{
			URI:         fmt.Sprintf("bank://accounts/%s/transactions", id),
			Name:        fmt.Sprintf("Transactions %s", id),
			Description: "Recent transactions",
		})
	}
	return out, nil
}

type readParams struct {
	URI string `json:"uri"`
}

func handleResourcesRead(bank *bankState, params json.RawMessage) (any, *rpcError) {
	var p readParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &rpcError{Code: -32602, Message: "Invalid params", Data: err.Error()}
	}
	if p.URI == "" {
		return nil, &rpcError{Code: -32602, Message: "Missing uri"}
	}

	bank.mu.Lock()
	defer bank.mu.Unlock()

	// bank://accounts
	if p.URI == "bank://accounts" {
		// Return a text block (MCP-style TextResourceContents)
		// You could also return structured JSON if your client expects it.
		var sb strings.Builder
		sb.WriteString("Accounts:\n")
		for _, a := range bank.accounts {
			sb.WriteString(fmt.Sprintf("- %s (%s) %s owner=%s\n", a.AccountID, a.Type, a.Currency, a.OwnerName))
		}
		return []TextResourceContents{{URI: p.URI, Text: sb.String()}}, nil
	}

	// bank://accounts/{id}
	if strings.HasPrefix(p.URI, "bank://accounts/") {
		rest := strings.TrimPrefix(p.URI, "bank://accounts/")
		parts := strings.Split(rest, "/")
		id := parts[0]
		acc, ok := bank.accounts[id]
		if !ok {
			return nil, &rpcError{Code: -32004, Message: "Account not found", Data: id}
		}

		if len(parts) == 1 {
			// details
			text := fmt.Sprintf("Account %s\nType: %s\nCurrency: %s\nOwner: %s\n",
				acc.AccountID, acc.Type, acc.Currency, acc.OwnerName)
			return []TextResourceContents{{URI: p.URI, Text: text}}, nil
		}

		switch parts[1] {
		case "balance":
			bal := bank.balances[id]
			text := fmt.Sprintf("Balance for %s\nAvailable: %.2f %s\nLedger: %.2f %s\n",
				id, bal.Available, bal.Currency, bal.Ledger, bal.Currency)
			return []TextResourceContents{{URI: p.URI, Text: text}}, nil
		case "transactions":
			txns := bank.transactions[id]
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Transactions for %s:\n", id))
			for _, t := range txns {
				sb.WriteString(fmt.Sprintf("- %s | %s | %s | %.2f %s | %s\n",
					t.TransactionID, t.Date, t.Description, t.Amount, t.Currency, t.Type))
			}
			return []TextResourceContents{{URI: p.URI, Text: sb.String()}}, nil
		default:
			return nil, &rpcError{Code: -32601, Message: "Unknown resource path", Data: p.URI}
		}
	}

	return nil, &rpcError{Code: -32601, Message: "Unknown resource", Data: p.URI}
}

// tools/call params
type toolCallParams struct {
	Name      string           `json:"name"`
	Arguments *json.RawMessage `json:"arguments,omitempty"`
}

func handleToolsCall(bank *bankState, params json.RawMessage) (any, *rpcError) {
	var p toolCallParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &rpcError{Code: -32602, Message: "Invalid params", Data: err.Error()}
	}
	switch p.Name {
	case "list_accounts":
		return toolListAccounts(bank)
	case "get_balance":
		return toolGetBalance(bank, p.Arguments)
	case "list_transactions":
		return toolListTransactions(bank, p.Arguments)
	case "create_transfer":
		return toolCreateTransfer(bank, p.Arguments)
	default:
		return nil, &rpcError{Code: -32601, Message: "Unknown tool", Data: p.Name}
	}
}

func toolListAccounts(bank *bankState) (any, *rpcError) {
	bank.mu.Lock()
	defer bank.mu.Unlock()
	var out []Account
	for _, a := range bank.accounts {
		out = append(out, a)
	}
	return out, nil
}

func toolGetBalance(bank *bankState, args *json.RawMessage) (any, *rpcError) {
	var p struct {
		AccountID string `json:"accountId"`
	}
	if args == nil || json.Unmarshal(*args, &p) != nil || p.AccountID == "" {
		return nil, &rpcError{Code: -32602, Message: "get_balance requires accountId"}
	}
	bank.mu.Lock()
	defer bank.mu.Unlock()
	bal, ok := bank.balances[p.AccountID]
	if !ok {
		return nil, &rpcError{Code: -32004, Message: "Account not found", Data: p.AccountID}
	}
	return bal, nil
}

func toolListTransactions(bank *bankState, args *json.RawMessage) (any, *rpcError) {
	var p struct {
		AccountID string `json:"accountId"`
		FromDate  string `json:"fromDate,omitempty"` // ISO date
		ToDate    string `json:"toDate,omitempty"`   // ISO date
	}
	if args == nil || json.Unmarshal(*args, &p) != nil || p.AccountID == "" {
		return nil, &rpcError{Code: -32602, Message: "list_transactions requires accountId"}
	}

	var from, to time.Time
	var err error
	if p.FromDate != "" {
		from, err = time.Parse("2006-01-02", p.FromDate)
		if err != nil {
			return nil, &rpcError{Code: -32602, Message: "Invalid fromDate", Data: p.FromDate}
		}
	}
	if p.ToDate != "" {
		to, err = time.Parse("2006-01-02", p.ToDate)
		if err != nil {
			return nil, &rpcError{Code: -32602, Message: "Invalid toDate", Data: p.ToDate}
		}
	}

	bank.mu.Lock()
	defer bank.mu.Unlock()
	txns, ok := bank.transactions[p.AccountID]
	if !ok {
		return nil, &rpcError{Code: -32004, Message: "Account not found", Data: p.AccountID}
	}

	// Filter by date range if provided (txn.Date is RFC3339)
	var out []Transaction
	for _, t := range txns {
		td, err := time.Parse(time.RFC3339, t.Date)
		if err != nil {
			continue
		}
		if !from.IsZero() && td.Before(from) {
			continue
		}
		if !to.IsZero() && td.After(to.Add(24*time.Hour)) { // inclusive toDate
			continue
		}
		out = append(out, t)
	}
	return out, nil
}

func toolCreateTransfer(bank *bankState, args *json.RawMessage) (any, *rpcError) {
	var req TransferRequest
	if args == nil || json.Unmarshal(*args, &req) != nil ||
		req.FromAccountID == "" || req.ToAccountID == "" || req.Amount <= 0 || req.Currency == "" {
		return nil, &rpcError{Code: -32602, Message: "create_transfer requires fromAccountId, toAccountId, amount>0, currency"}
	}

	bank.mu.Lock()
	defer bank.mu.Unlock()

	fromBal, ok := bank.balances[req.FromAccountID]
	if !ok {
		return nil, &rpcError{Code: -32004, Message: "From account not found", Data: req.FromAccountID}
	}
	toBal, ok := bank.balances[req.ToAccountID]
	if !ok {
		return nil, &rpcError{Code: -32004, Message: "To account not found", Data: req.ToAccountID}
	}
	if fromBal.Currency != req.Currency || toBal.Currency != req.Currency {
		return nil, &rpcError{Code: -32002, Message: "Currency mismatch"}
	}
	if fromBal.Available < req.Amount {
		return nil, &rpcError{Code: -32001, Message: "Insufficient funds"}
	}

	// Apply transfer
	now := time.Now().UTC().Format(time.RFC3339)
	transferID := fmt.Sprintf("tr-%d", time.Now().UnixNano())

	fromBal.Available -= req.Amount
	fromBal.Ledger -= req.Amount
	toBal.Available += req.Amount
	toBal.Ledger += req.Amount
	bank.balances[req.FromAccountID] = fromBal
	bank.balances[req.ToAccountID] = toBal

	// Record transactions
	fromTxn := Transaction{
		TransactionID: "tx-out-" + transferID,
		Date:          now,
		Description:   nonEmpty(req.Description, "Transfer out"),
		Amount:        -req.Amount,
		Currency:      req.Currency,
		Type:          "debit",
	}
	toTxn := Transaction{
		TransactionID: "tx-in-" + transferID,
		Date:          now,
		Description:   nonEmpty(req.Description, "Transfer in"),
		Amount:        req.Amount,
		Currency:      req.Currency,
		Type:          "credit",
	}
	bank.transactions[req.FromAccountID] = append(bank.transactions[req.FromAccountID], fromTxn)
	bank.transactions[req.ToAccountID] = append(bank.transactions[req.ToAccountID], toTxn)

	return TransferResponse{
		TransferID: transferID,
		Status:     "completed",
		Timestamp:  now,
	}, nil
}

func nonEmpty(s, fallback string) string {
	if strings.TrimSpace(s) == "" {
		return fallback
	}
	return s
}

// ---------- Server Loop ----------
func main() {
	bank := newBank()
	reader := bufio.NewReader(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return
			}
			// On read error, try to report and continue
			_ = encoder.Encode(errResult(nil, -32700, "Read error", err.Error()))
			continue
		}
		line = bytesTrimSpace(line)
		if len(line) == 0 {
			continue
		}

		var req rpcRequest
		if err := json.Unmarshal(line, &req); err != nil || req.JSONRPC != jsonrpcVersion {
			_ = encoder.Encode(errResult(nil, -32700, "Parse error", string(line)))
			continue
		}

		var (
			result any
			rpcErr *rpcError
		)

		switch req.Method {
		case "ping":
			result, rpcErr = handlePing(req.Params)
		case "resources/list":
			result, rpcErr = handleResourcesList(bank, req.Params)
		case "resources/read":
			result, rpcErr = handleResourcesRead(bank, req.Params)
		case "tools/call":
			result, rpcErr = handleToolsCall(bank, req.Params)
		default:
			rpcErr = &rpcError{Code: -32601, Message: "Method not found", Data: req.Method}
		}

		if rpcErr != nil {
			_ = encoder.Encode(errResult(req.ID, rpcErr.Code, rpcErr.Message, rpcErr.Data))
		} else {
			_ = encoder.Encode(okResult(req.ID, result))
		}
	}
}

// bytesTrimSpace is a small helper avoiding importing bytes for just TrimSpace on []byte.
func bytesTrimSpace(b []byte) []byte {
	start := 0
	for start < len(b) && (b[start] == ' ' || b[start] == '\t' || b[start] == '\r' || b[start] == '\n') {
		start++
	}
	end := len(b) - 1
	for end >= start && (b[end] == ' ' || b[end] == '\t' || b[end] == '\r' || b[end] == '\n') {
		end--
	}
	return b[start : end+1]
}
