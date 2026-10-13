package cli

import (
	"context"
	"errors"
	"path/filepath"
	"sync/atomic"
	"testing"
	"testing/synctest"

	"github.com/ivuorinen/gibidify/fileproc"
	"github.com/ivuorinen/gibidify/testutil"
)

// cancelAfterFirstCheck is a context that reports itself live to the first Err
// call and canceled to every later one, so a cancellation can land between
// processFile's own check and the file processor's result.
type cancelAfterFirstCheck struct {
	context.Context
	calls atomic.Int32
}

func (c *cancelAfterFirstCheck) Err() error {
	if c.calls.Add(1) == 1 {
		return nil
	}

	return context.Canceled
}

// TestProcessFileOutcomes verifies how processFile counts each outcome. A
// missing file would count as failed, so the canceled cases show cancellation
// is neither attempted nor counted.
func TestProcessFileOutcomes(t *testing.T) {
	testutil.SetConfigKeys(t, nil)
	dir := t.TempDir()
	binary := testutil.CreateTestFile(t, dir, "blob.dat", []byte{0x00, 0x01, 0x02, 0x00, 0xff})
	missing := filepath.Join(dir, "missing.go")
	canceled, cancel := context.WithCancel(context.Background())
	cancel()

	tests := []struct {
		name        string
		ctx         context.Context
		path        string
		wantSkipped int64
	}{
		{"canceled before start", canceled, missing, 0},
		{"canceled while processing", &cancelAfterFirstCheck{Context: context.Background()}, missing, 0},
		{"binary content skipped", context.Background(), binary, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := &runStats{}
			job := fileJob{path: tt.path, out: make(chan fileproc.WriteRequest, 1)}
			NewProcessor(pipelineFlags(dir, "", 1)).processFile(tt.ctx, dir, job, stats)

			if _, sent := <-job.out; sent {
				t.Error("processFile sent a WriteRequest, want out closed empty")
			}
			if got := stats.skipped.Load(); got != tt.wantSkipped {
				t.Errorf("skipped = %d, want %d", got, tt.wantSkipped)
			}
			if got := stats.failed.Load(); got != 0 {
				t.Errorf("failed = %d, want 0", got)
			}
		})
	}
}

// TestSendFilesCanceledWhileBlocked verifies sendFiles returns on cancellation
// while blocked handing a job to the workers or queueing its result channel,
// and closes both channels.
func TestSendFilesCanceledWhileBlocked(t *testing.T) {
	tests := []struct {
		name    string
		jobsCap int
	}{
		{"blocked on jobs", 0},
		{"blocked on order", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(t.Context())
				jobs := make(chan fileJob, tt.jobsCap)
				order := make(chan chan fileproc.WriteRequest)
				done := make(chan error, 1)
				p := NewProcessor(&Flags{NoUI: true})
				go func() { done <- p.sendFiles(ctx, []string{"a.go", "b.go"}, jobs, order) }()

				synctest.Wait()
				cancel()
				if err := <-done; !errors.Is(err, context.Canceled) {
					t.Errorf("sendFiles() error = %v, want context.Canceled", err)
				}
				assertSentJobs(t, jobs, order, tt.jobsCap)
			})
		})
	}
}

// assertSentJobs drains jobs and order, which sendFiles must have closed, and
// checks wantJobs jobs reached the workers and no result channel was queued.
func assertSentJobs(t *testing.T, jobs <-chan fileJob, order <-chan chan fileproc.WriteRequest, wantJobs int) {
	t.Helper()
	gotJobs := 0
	for range jobs {
		gotJobs++
	}
	if gotJobs != wantJobs {
		t.Errorf("jobs sent = %d, want %d", gotJobs, wantJobs)
	}
	if _, ok := <-order; ok {
		t.Error("order queued a result channel, want it closed empty")
	}
}
