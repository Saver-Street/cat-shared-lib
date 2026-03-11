package schedule_test

import (
	"context"
	"fmt"
	"time"

	"github.com/Saver-Street/cat-shared-lib/schedule"
)

func ExampleScheduler() {
	s := schedule.New()
	defer s.Stop()

	s.Every("heartbeat", 100*time.Millisecond, func(_ context.Context) {
		fmt.Println("tick")
	})

	time.Sleep(250 * time.Millisecond)
	s.Stop()
	// Output:
	// tick
	// tick
}
