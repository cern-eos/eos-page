package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	openAICreditGrantsURL = "https://api.openai.com/v1/dashboard/billing/credit_grants"
	openAISpendLimitURL   = "https://api.openai.com/v1/organization/spend_limit"
	openAICostsURL        = "https://api.openai.com/v1/organization/costs"
)

type Budget struct {
	OK        bool    `json:"ok"`
	Available float64 `json:"available"`
	Used      float64 `json:"used"`
	Limit     float64 `json:"limit"`
	HasAvail  bool    `json:"hasAvailable"`
	HasUsed   bool    `json:"hasUsed"`
	HasLimit  bool    `json:"hasLimit"`
	Currency  string  `json:"currency"`
	Source    string  `json:"source"`
	Note      string  `json:"note,omitempty"`
	FetchedAt string  `json:"fetchedAt"`
}

func FetchBudget(ctx context.Context, apiKey, adminKey string, monthlyBudgetUSD float64) Budget {
	out := Budget{Currency: "USD", FetchedAt: time.Now().UTC().Format(time.RFC3339)}
	apiKey = strings.TrimSpace(apiKey)
	adminKey = strings.TrimSpace(adminKey)
	if apiKey == "" && adminKey == "" {
		out.Note = "OPENAI_API_KEY is not set"
		return out
	}

	if grants, err := fetchCreditGrants(ctx, firstKey(apiKey, adminKey)); err == nil && (grants.HasAvail || grants.HasUsed) {
		grants.FetchedAt = out.FetchedAt
		return grants
	}

	key := firstKey(adminKey, apiKey)
	limit, hasLimit, limitErr := fetchSpendLimitUSD(ctx, key)
	used, hasUsed, usedErr := fetchMonthCostUSD(ctx, key)
	if hasLimit {
		out.Limit = limit
		out.HasLimit = true
	} else if monthlyBudgetUSD > 0 {
		out.Limit = monthlyBudgetUSD
		out.HasLimit = true
		out.Source = "OPENAI_BUDGET_USD"
	}
	if hasUsed {
		out.Used = used
		out.HasUsed = true
	}
	if out.HasLimit && out.HasUsed {
		out.Available = out.Limit - out.Used
		if out.Available < 0 {
			out.Available = 0
		}
		out.HasAvail = true
		out.OK = true
		if out.Source == "" {
			out.Source = "organization/spend_limit + costs"
		} else {
			out.Source += " + organization/costs"
		}
		return out
	}
	if out.HasUsed {
		out.OK = true
		out.Source = "organization/costs"
		out.Note = "Month spend only. Set OPENAI_ADMIN_KEY or OPENAI_BUDGET_USD to show remaining budget."
		return out
	}

	out.Note = budgetFetchNote(limitErr, usedErr)
	return out
}

func firstKey(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return strings.TrimSpace(a)
	}
	return strings.TrimSpace(b)
}

func budgetFetchNote(limitErr, usedErr error) string {
	for _, err := range []error{usedErr, limitErr} {
		if err == nil {
			continue
		}
		msg := err.Error()
		if strings.Contains(msg, "session key") || strings.Contains(msg, "403") || strings.Contains(strings.ToLower(msg), "admin") {
			return "OpenAI remaining budget needs an organization admin key (OPENAI_ADMIN_KEY)."
		}
		return msg
	}
	return "Could not read OpenAI budget."
}

func fetchCreditGrants(ctx context.Context, key string) (Budget, error) {
	raw, err := openAIGet(ctx, key, openAICreditGrantsURL)
	if err != nil {
		return Budget{}, err
	}
	var body struct {
		TotalGranted   float64 `json:"total_granted"`
		TotalUsed      float64 `json:"total_used"`
		TotalAvailable float64 `json:"total_available"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return Budget{}, err
	}
	out := Budget{
		OK:        true,
		Available: body.TotalAvailable,
		Used:      body.TotalUsed,
		Limit:     body.TotalGranted,
		HasAvail:  true,
		HasUsed:   true,
		HasLimit:  body.TotalGranted > 0,
		Currency:  "USD",
		Source:    "dashboard/billing/credit_grants",
	}
	return out, nil
}

func fetchSpendLimitUSD(ctx context.Context, key string) (float64, bool, error) {
	raw, err := openAIGet(ctx, key, openAISpendLimitURL)
	if err != nil {
		return 0, false, err
	}
	var body struct {
		ThresholdAmount float64 `json:"threshold_amount"`
		Currency        string  `json:"currency"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return 0, false, err
	}
	if body.ThresholdAmount <= 0 {
		return 0, false, nil
	}
	return body.ThresholdAmount / 100, true, nil
}

func fetchMonthCostUSD(ctx context.Context, key string) (float64, bool, error) {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	u, err := url.Parse(openAICostsURL)
	if err != nil {
		return 0, false, err
	}
	q := u.Query()
	q.Set("start_time", strconv.FormatInt(start.Unix(), 10))
	q.Set("end_time", strconv.FormatInt(now.Add(24*time.Hour).Unix(), 10))
	q.Set("bucket_width", "1d")
	q.Set("limit", "31")
	u.RawQuery = q.Encode()
	raw, err := openAIGet(ctx, key, u.String())
	if err != nil {
		return 0, false, err
	}
	sum, ok := sumCostUSD(raw)
	return sum, ok, nil
}

func sumCostUSD(raw []byte) (float64, bool) {
	var body struct {
		Data []struct {
			Results []struct {
				Amount *struct {
					Value    float64 `json:"value"`
					Currency string  `json:"currency"`
				} `json:"amount"`
			} `json:"results"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return 0, false
	}
	var sum float64
	found := false
	for _, bucket := range body.Data {
		for _, row := range bucket.Results {
			if row.Amount == nil {
				continue
			}
			sum += row.Amount.Value
			found = true
		}
	}
	return sum, found
}

func openAIGet(ctx context.Context, key, rawURL string) ([]byte, error) {
	if key == "" {
		return nil, fmt.Errorf("missing openai key")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Accept", "application/json")
	res, err := openaiHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 300 {
		var wrap struct {
			Error *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.Unmarshal(raw, &wrap)
		msg := strings.TrimSpace(string(raw))
		if wrap.Error != nil && wrap.Error.Message != "" {
			msg = wrap.Error.Message
		}
		return nil, fmt.Errorf("openai %s: %s", res.Status, msg)
	}
	return raw, nil
}
