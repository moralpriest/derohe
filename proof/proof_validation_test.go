package proof

import "testing"

func TestValidatePayloadProofAmount(t *testing.T) {
	tests := []struct {
		name        string
		amount      uint64
		shouldError bool
	}{
		{
			name:        "normal amount (100 DERO)",
			amount:      10_000_000, // 100 DERO
			shouldError: false,
		},
		{
			name:        "large legitimate amount (1M DERO)",
			amount:      100_000_000_000, // 1M DERO
			shouldError: false,
		},
		{
			name:        "very large but valid (10M DERO)",
			amount:      1_000_000_000_000, // 10M DERO
			shouldError: false,
		},
		{
			name:        "at current supply (16.5M DERO)",
			amount:      16_500_000_000_000, // 16.5M DERO - current supply
			shouldError: false,
		},
		{
			name:        "near hard cap (21M DERO)",
			amount:      21_000_000_000_000, // 21M DERO - hard cap
			shouldError: false,
		},
		{
			name:        "exceeds hard cap (23M DERO)",
			amount:      23_000_000_000_000, // 23M DERO - above 22M threshold
			shouldError: true,
		},
		{
			name:        "exceeds hard cap (50M DERO)",
			amount:      50_000_000_000_000, // 50M DERO
			shouldError: true,
		},
		{
			name:        "known impossible amount (184 trillion DERO)",
			amount:      18_446_743_673_709_551_434,
			shouldError: true,
		},
		{
			name:        "wraparound amount",
			amount:      18_446_744_073_707_351_616,
			shouldError: true,
		},
		{
			name:        "maximum safe int64",
			amount:      9_223_372_036_854_775_807,
			shouldError: true,
		},
		{
			name:        "just above int64 max",
			amount:      9_223_372_036_854_775_808,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePayloadProofAmount(tt.amount)
			if tt.shouldError && err == nil {
				t.Fatalf("expected amount %d to be rejected", tt.amount)
			}
			if !tt.shouldError && err != nil {
				t.Fatalf("expected amount %d to be accepted: %v", tt.amount, err)
			}
		})
	}
}

func TestKnownFakeProofAmounts(t *testing.T) {
	fakeProofs := []struct {
		name   string
		amount uint64
	}{
		{
			name:   "demonstrated impossible amount",
			amount: 18_446_743_673_709_551_434,
		},
		{
			name:   "reported impossible amount",
			amount: 18_446_743_853_709_551_434,
		},
	}

	for _, fake := range fakeProofs {
		t.Run(fake.name, func(t *testing.T) {
			if err := ValidatePayloadProofAmount(fake.amount); err == nil {
				t.Fatalf("expected fake proof amount %d to be rejected", fake.amount)
			}
		})
	}
}
