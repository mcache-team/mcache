package cluster

import (
	"strconv"
	"strings"
	"time"
)

func BuildMetrics(diag *Diagnostics) string {
	var builder strings.Builder

	writeMetricLine(&builder, "mcache_state_items_total", nil, float64(diag.State.ItemCount))
	writeMetricLine(&builder, "mcache_state_roots_total", nil, float64(diag.State.RootCount))
	writeMetricLine(&builder, "mcache_cluster_ready", nil, boolToFloat(diag.Ready))
	writeMetricLine(&builder, "mcache_cluster_is_leader", nil, boolToFloat(diag.Status.IsLeader))
	writeMetricLine(&builder, "mcache_cluster_members_total", nil, float64(len(diag.Members)))
	writeMetricLine(&builder, "mcache_cluster_role_info", map[string]string{
		"mode":    diag.Status.Mode,
		"node_id": diag.Status.NodeID,
		"role":    diag.Status.Role,
	}, 1)

	for operation, stat := range diag.Telemetry.Writes {
		labels := map[string]string{"operation": operation}
		writeMetricLine(&builder, "mcache_write_requests_total", labels, float64(stat.Requests))
		writeMetricLine(&builder, "mcache_write_success_total", labels, float64(stat.Successes))
		writeMetricLine(&builder, "mcache_write_error_total", labels, float64(stat.Errors))
		writeMetricLine(&builder, "mcache_write_redirect_total", labels, float64(stat.Redirects))
		writeMetricLine(&builder, "mcache_write_latency_seconds_total", labels, float64(stat.LatencyTotalNs)/float64(time.Second))
		writeMetricLine(&builder, "mcache_write_latency_last_seconds", labels, float64(stat.LastLatencyNs)/float64(time.Second))
	}

	for _, metric := range []struct {
		raftKey   string
		metricKey string
	}{
		{raftKey: "term", metricKey: "mcache_raft_term"},
		{raftKey: "last_log_index", metricKey: "mcache_raft_last_log_index"},
		{raftKey: "commit_index", metricKey: "mcache_raft_commit_index"},
		{raftKey: "applied_index", metricKey: "mcache_raft_applied_index"},
		{raftKey: "fsm_pending", metricKey: "mcache_raft_fsm_pending"},
		{raftKey: "last_snapshot_index", metricKey: "mcache_raft_last_snapshot_index"},
		{raftKey: "latest_configuration_index", metricKey: "mcache_raft_latest_configuration_index"},
	} {
		if diag.Raft == nil {
			continue
		}
		rawValue, ok := diag.Raft[metric.raftKey]
		if !ok {
			continue
		}
		value, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			continue
		}
		writeMetricLine(&builder, metric.metricKey, nil, value)
	}

	return builder.String()
}

func writeMetricLine(builder *strings.Builder, name string, labels map[string]string, value float64) {
	builder.WriteString(name)
	if len(labels) > 0 {
		builder.WriteString("{")
		first := true
		for key, labelValue := range labels {
			if !first {
				builder.WriteString(",")
			}
			first = false
			builder.WriteString(key)
			builder.WriteString("=\"")
			builder.WriteString(escapeLabelValue(labelValue))
			builder.WriteString("\"")
		}
		builder.WriteString("}")
	}
	builder.WriteString(" ")
	builder.WriteString(strconv.FormatFloat(value, 'f', -1, 64))
	builder.WriteString("\n")
}

func escapeLabelValue(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, "\n", `\n`, `"`, `\"`)
	return replacer.Replace(value)
}

func boolToFloat(value bool) float64 {
	if value {
		return 1
	}
	return 0
}
