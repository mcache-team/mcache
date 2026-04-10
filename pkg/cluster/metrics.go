package cluster

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

func BuildMetrics(diag *Diagnostics) string {
	var builder strings.Builder

	writeMetricMetadata(&builder, "mcache_state_items_total", "gauge", "Current number of cached items in the local state machine.")
	writeMetricMetadata(&builder, "mcache_state_roots_total", "gauge", "Current number of root prefixes in the local state machine.")
	writeMetricMetadata(&builder, "mcache_cluster_ready", "gauge", "Whether this node is currently ready to serve cluster traffic.")
	writeMetricMetadata(&builder, "mcache_cluster_is_leader", "gauge", "Whether this node is currently the raft leader.")
	writeMetricMetadata(&builder, "mcache_cluster_members_total", "gauge", "Current number of known cluster members.")
	writeMetricMetadata(&builder, "mcache_cluster_role_info", "gauge", "Static info metric for the current node role and mode.")
	writeMetricMetadata(&builder, "mcache_write_requests_total", "counter", "Total number of observed write requests by operation.")
	writeMetricMetadata(&builder, "mcache_write_success_total", "counter", "Total number of successful write requests by operation.")
	writeMetricMetadata(&builder, "mcache_write_error_total", "counter", "Total number of failed write requests by operation.")
	writeMetricMetadata(&builder, "mcache_write_redirect_total", "counter", "Total number of write requests redirected to another leader.")
	writeMetricMetadata(&builder, "mcache_write_latency_last_seconds", "gauge", "Latency in seconds of the last observed write request by operation.")
	writeMetricMetadata(&builder, "mcache_write_latency_seconds", "histogram", "Write request latency histogram in seconds by operation.")
	writeMetricMetadata(&builder, "mcache_raft_term", "gauge", "Current raft term for this node.")
	writeMetricMetadata(&builder, "mcache_raft_last_log_index", "gauge", "Last raft log index known locally.")
	writeMetricMetadata(&builder, "mcache_raft_commit_index", "gauge", "Current committed raft log index.")
	writeMetricMetadata(&builder, "mcache_raft_applied_index", "gauge", "Current applied raft log index.")
	writeMetricMetadata(&builder, "mcache_raft_fsm_pending", "gauge", "Number of pending FSM mutations queued in raft.")
	writeMetricMetadata(&builder, "mcache_raft_last_snapshot_index", "gauge", "Last raft snapshot index known locally.")
	writeMetricMetadata(&builder, "mcache_raft_latest_configuration_index", "gauge", "Latest raft configuration log index.")

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
		writeMetricLine(&builder, "mcache_write_latency_last_seconds", labels, float64(stat.LastLatencyNs)/float64(time.Second))
		writeMetricLine(&builder, "mcache_write_latency_seconds_sum", labels, float64(stat.LatencyTotalNs)/float64(time.Second))
		writeMetricLine(&builder, "mcache_write_latency_seconds_count", labels, float64(stat.Requests))
		for index, bucket := range writeLatencyBuckets {
			bucketLabels := cloneLabels(labels)
			bucketLabels["le"] = formatDurationSeconds(bucket)
			writeMetricLine(&builder, "mcache_write_latency_seconds_bucket", bucketLabels, float64(stat.LatencyBucketCounts[index]))
		}
		infLabels := cloneLabels(labels)
		infLabels["le"] = "+Inf"
		writeMetricLine(&builder, "mcache_write_latency_seconds_bucket", infLabels, float64(stat.LatencyInfBucket))
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

func writeMetricMetadata(builder *strings.Builder, name, metricType, help string) {
	builder.WriteString("# HELP ")
	builder.WriteString(name)
	builder.WriteString(" ")
	builder.WriteString(help)
	builder.WriteString("\n")
	builder.WriteString("# TYPE ")
	builder.WriteString(name)
	builder.WriteString(" ")
	builder.WriteString(metricType)
	builder.WriteString("\n")
}

func writeMetricLine(builder *strings.Builder, name string, labels map[string]string, value float64) {
	builder.WriteString(name)
	if len(labels) > 0 {
		builder.WriteString("{")
		keys := make([]string, 0, len(labels))
		for key := range labels {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for index, key := range keys {
			if index > 0 {
				builder.WriteString(",")
			}
			builder.WriteString(key)
			builder.WriteString("=\"")
			builder.WriteString(escapeLabelValue(labels[key]))
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

func cloneLabels(labels map[string]string) map[string]string {
	cloned := make(map[string]string, len(labels))
	for key, value := range labels {
		cloned[key] = value
	}
	return cloned
}

func formatDurationSeconds(value time.Duration) string {
	return strconv.FormatFloat(value.Seconds(), 'f', -1, 64)
}

func boolToFloat(value bool) float64 {
	if value {
		return 1
	}
	return 0
}
