package bandwidth

import "testing"

// TestNewRateConfig tests the validations we run like setting burst = limit if burst <= 0
func TestNewRateConfig(t *testing.T) {

	conf := NewRateConfig(10, 0)

	if conf.Burst() != 10 {
		t.Errorf("expected burst = limit but got %d", conf.Burst())
	}

	conf.SetBurst(5)
	if conf.Burst() != 5 {
		t.Fatalf("expected burst = 5 but got %d", conf.Burst())
	}

	conf.SetLimit(100)
	conf.SetBurst(0)

	if int(conf.Limit()) != conf.Burst() {
		t.Fatalf("expected burst = %d, but got %d", conf.Limit(), conf.Burst())
	}
}
