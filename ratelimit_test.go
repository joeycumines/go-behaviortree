package behaviortree

import (
	"testing"
	"time"
)

func TestRateLimit(t *testing.T) {
	var (
		count    = 0
		duration = time.Millisecond * 100
		node     = New(
			Sequence,
			New(RateLimit(duration)),
			New(func(children []Node) (Status, error) {
				count++
				return Success, nil
			}),
		)
	)
	timer := time.NewTimer(duration*4 + (duration / 2))
	defer timer.Stop()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if _, err := node.Tick(); err != nil {
				t.Fatal(err)
			}
			continue
		case <-timer.C:
		}
		if count != 5 {
			t.Fatal(count)
		}
		return
	}
}
