package chat

import "testing"

func TestSumCostUSD(t *testing.T) {
	raw := []byte(`{"data":[
		{"results":[{"amount":{"value":1.25,"currency":"usd"}}]},
		{"results":[{"amount":{"value":0.75,"currency":"usd"}}]}
	]}`)
	sum, ok := sumCostUSD(raw)
	if !ok || sum != 2.0 {
		t.Fatalf("sum=%v ok=%v", sum, ok)
	}
}

func TestFetchBudgetNoKeys(t *testing.T) {
	b := FetchBudget(t.Context(), "", "", 0)
	if b.OK || b.Note == "" {
		t.Fatalf("%+v", b)
	}
}
