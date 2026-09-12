package main

import (
	"context"
	"strings"
	"sync"
)

func runDetailJobs(ctx context.Context, count, limit int, run func(int) ([]map[string]interface{}, float64, string)) ([]map[string]interface{}, float64, string) {
	limit = max(1, min(limit, count))
	groups := make([][]map[string]interface{}, count)
	costs := make([]float64, count)
	messages := make([]string, count)
	semaphore := make(chan struct{}, limit)
	var workers sync.WaitGroup
	for i := 0; i < count; i++ {
		workers.Add(1)
		go func(index int) {
			defer workers.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			if err := ctx.Err(); err != nil {
				messages[index] = err.Error()
				return
			}
			groups[index], costs[index], messages[index] = run(index)
		}(i)
	}
	workers.Wait()
	var items []map[string]interface{}
	var total float64
	var failures []string
	for i := range groups {
		items = append(items, groups[i]...)
		total += costs[i]
		if messages[i] != "" {
			failures = append(failures, messages[i])
		}
	}
	return items, total, strings.Join(failures, "；")
}
