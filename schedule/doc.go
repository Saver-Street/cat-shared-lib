// Package schedule provides a lightweight in-process task scheduler for
// running periodic functions at fixed intervals.
//
// [Scheduler] manages a set of named tasks, each running on its own ticker.
// Tasks start automatically when added and run until the scheduler is stopped.
//
//	s := schedule.New()
//	s.Every("cleanup-stale", 5*time.Minute, func(ctx context.Context) {
//	    registry.MarkStale(10 * time.Minute)
//	})
//	defer s.Stop()
//
// Tasks respect the context passed to [Scheduler.Stop] and stop gracefully.
// Use [Scheduler.Remove] to cancel a single task.
package schedule
