package cluster

import (
	"fmt"
	"sort"
	"strings"
)

type Peer struct {
	NodeID        string
	RaftAddress   string
	AdvertiseAddr string
}

func parsePeers(value string) ([]Peer, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	rawPeers := strings.Split(value, ",")
	peers := make([]Peer, 0, len(rawPeers))
	seen := make(map[string]struct{}, len(rawPeers))
	for _, rawPeer := range rawPeers {
		rawPeer = strings.TrimSpace(rawPeer)
		if rawPeer == "" {
			continue
		}
		parts := strings.Split(rawPeer, "@")
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid peer %q, expected nodeID@raftAddr@httpAddr", rawPeer)
		}
		peer := Peer{
			NodeID:        strings.TrimSpace(parts[0]),
			RaftAddress:   strings.TrimSpace(parts[1]),
			AdvertiseAddr: strings.TrimSpace(parts[2]),
		}
		if peer.NodeID == "" || peer.RaftAddress == "" || peer.AdvertiseAddr == "" {
			return nil, fmt.Errorf("invalid peer %q, empty fields are not allowed", rawPeer)
		}
		if _, ok := seen[peer.NodeID]; ok {
			return nil, fmt.Errorf("duplicated peer node id %q", peer.NodeID)
		}
		seen[peer.NodeID] = struct{}{}
		peers = append(peers, peer)
	}
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].NodeID < peers[j].NodeID
	})
	return peers, nil
}
