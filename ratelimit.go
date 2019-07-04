package behaviortree

import "time"

// RateLimit generates a stateful Tick that will return success at most once per a given duration
func RateLimit(d time.Duration) Tick {
	var last *time.Time
	return func(children []Node) (Status, error) {
		now := time.Now()
		if last != nil && now.Add(-d).Before(*last) {
			return Failure, nil
		}
		last = &now
		return Success, nil
	}
}
