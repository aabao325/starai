package main

import (
	"context"
	"testing"
	"time"
)

func TestRunDetailJobsPreservesModuleOrder(t *testing.T) {
	items, cost, message := runDetailJobs(context.Background(), 3, 2, func(index int) ([]map[string]interface{}, float64, string) {
		time.Sleep(time.Duration(3-index) * time.Millisecond)
		failure := ""
		if index == 1 {
			failure = "module failed"
		}
		return []map[string]interface{}{{"index": index}}, float64(index + 1), failure
	})
	if len(items) != 3 || intAny(items[0]["index"]) != 0 || intAny(items[1]["index"]) != 1 || intAny(items[2]["index"]) != 2 {
		t.Fatalf("items out of order: %#v", items)
	}
	if cost != 6 || message != "module failed" {
		t.Fatalf("cost=%v message=%q", cost, message)
	}
}
