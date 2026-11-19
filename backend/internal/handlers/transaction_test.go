package handlers

import (
	"encoding/json"
	"testing"
)

func TestCreateTransactionRequest_AutoTipFor(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		clubOn  bool
		wantTip bool
	}{
		{"club enabled, no opt-out", `{"transaction_type":"deposit","amount":100}`, true, true},
		{"club enabled, opt-out", `{"transaction_type":"deposit","amount":100,"skip_auto_tip":true}`, true, false},
		{"club enabled, explicit false", `{"transaction_type":"deposit","amount":100,"skip_auto_tip":false}`, true, true},
		{"club disabled", `{"transaction_type":"deposit","amount":100}`, false, false},
		{"club disabled, opt-out", `{"transaction_type":"deposit","amount":100,"skip_auto_tip":true}`, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req createTransactionRequest
			if err := json.Unmarshal([]byte(tt.body), &req); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got := req.autoTipFor(tt.clubOn); got != tt.wantTip {
				t.Errorf("autoTipFor(%v) = %v, want %v", tt.clubOn, got, tt.wantTip)
			}
		})
	}
}
