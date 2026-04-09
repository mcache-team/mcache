package cluster

import "fmt"

type NotLeaderError struct {
	LeaderAddress string
}

func (e *NotLeaderError) Error() string {
	if e.LeaderAddress == "" {
		return "request must be sent to cluster leader"
	}
	return fmt.Sprintf("request must be sent to cluster leader at %s", e.LeaderAddress)
}
