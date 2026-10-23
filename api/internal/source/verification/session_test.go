package verification

import (
	"context"
	"errors"
	"github.com/yueli-official/ytv/api/internal/source"
	"sync/atomic"
	"testing"
	"time"
)

type testFlow struct {
	closes        atomic.Int32
	replays       atomic.Int32
	submissions   atomic.Int32
	fail          error
	entered       chan struct{}
	unblock       chan struct{}
	resumeEntered chan struct{}
	resumeUnblock chan struct{}
}

func (f *testFlow) Refresh(context.Context) (CaptchaImage, error) {
	return CaptchaImage{MIME: "image/png", Bytes: []byte{1}}, nil
}
func (f *testFlow) Submit(ctx context.Context, answer string) (bool, error) {
	f.submissions.Add(1)
	if f.entered != nil {
		close(f.entered)
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-f.unblock:
		}
	}
	return answer == "right", f.fail
}
func (f *testFlow) Resume(context.Context) (source.Page, error) {
	f.replays.Add(1)
	if f.resumeEntered != nil {
		close(f.resumeEntered)
		<-f.resumeUnblock
	}
	return source.Page{Items: []source.Title{{Name: "original query result"}}}, f.fail
}
func (f *testFlow) Close() { f.closes.Add(1) }

func TestShutdownCanCancelAlreadyAcceptedResume(t *testing.T) {
	s := NewSessions()
	ctx := context.Background()
	f := &testFlow{resumeEntered: make(chan struct{}), resumeUnblock: make(chan struct{})}
	c, _ := s.Start(ctx, "alice", f)
	result := make(chan error, 1)
	go func() { _, err := s.Submit(ctx, "alice", c.ID, "right"); result <- err }()
	<-f.resumeEntered
	if _, err := s.Submit(ctx, "alice", c.ID, "right"); !errors.Is(err, ErrSession) {
		t.Fatal("consumed session reused")
	}
	s.Close()
	if f.closes.Load() != 1 {
		t.Fatal("shutdown lost active resume")
	}
	close(f.resumeUnblock)
	if err := <-result; !errors.Is(err, ErrSession) || f.closes.Load() != 1 {
		t.Fatal("closed session published result")
	}
}

func TestSessionOwnershipRetryAndSingleResume(t *testing.T) {
	s := NewSessions()
	defer s.Close()
	ctx := context.Background()
	f := &testFlow{}
	challenge, err := s.Start(ctx, "alice", f)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Submit(ctx, "bob", challenge.ID, "right"); !errors.Is(err, ErrSession) {
		t.Fatal("cross-user submit accepted")
	}
	if _, err = s.Refresh(ctx, "bob", challenge.ID); !errors.Is(err, ErrSession) {
		t.Fatal("cross-user refresh accepted")
	}
	if err = s.Cancel("bob", challenge.ID); !errors.Is(err, ErrSession) {
		t.Fatal("cross-user cancel accepted")
	}
	bad, err := s.Submit(ctx, "alice", challenge.ID, "wrong")
	if err != nil || bad.Accepted || f.replays.Load() != 0 {
		t.Fatal("incorrect answer replayed")
	}
	if _, err = s.Refresh(ctx, "alice", challenge.ID); err != nil {
		t.Fatal(err)
	}
	good, err := s.Submit(ctx, "alice", challenge.ID, "right")
	if err != nil || !good.Accepted || len(good.Page.Items) != 1 {
		t.Fatalf("resume: %#v %v", good, err)
	}
	if _, err = s.Submit(ctx, "alice", challenge.ID, "right"); !errors.Is(err, ErrSession) {
		t.Fatal("duplicate accepted")
	}
	if f.replays.Load() != 1 || f.closes.Load() != 1 {
		t.Fatal("flow not consumed exactly once")
	}
}

func TestSessionExpiryCancelCapacityAndShutdown(t *testing.T) {
	s := NewSessions()
	now := time.Now()
	s.clock = func() time.Time { return now }
	s.perOwner = 1
	s.limit = 2
	ctx := context.Background()
	f := &testFlow{}
	c, _ := s.Start(ctx, "alice", f)
	extra := &testFlow{}
	if _, err := s.Start(ctx, "alice", extra); !errors.Is(err, ErrLimit) || extra.closes.Load() != 1 {
		t.Fatal("owner cap")
	}
	second, _ := s.Start(ctx, "bob", &testFlow{})
	if _, err := s.Start(ctx, "charlie", &testFlow{}); !errors.Is(err, ErrLimit) {
		t.Fatal("global cap")
	}
	if err := s.Cancel("bob", second.ID); err != nil {
		t.Fatal(err)
	}
	now = now.Add(s.ttl)
	if _, err := s.Submit(ctx, "alice", c.ID, "right"); !errors.Is(err, ErrSession) || f.submissions.Load() != 0 || f.closes.Load() != 1 {
		t.Fatal("expired session used")
	}
	last := &testFlow{}
	s.Start(ctx, "alice", last)
	s.Close()
	if last.closes.Load() != 1 {
		t.Fatal("shutdown leaked session")
	}
	if _, err := s.Start(ctx, "alice", &testFlow{}); !errors.Is(err, ErrLimit) {
		t.Fatal("start after shutdown")
	}
}

func TestBusyAndCancelledSubmissionCannotResume(t *testing.T) {
	s := NewSessions()
	defer s.Close()
	ctx := context.Background()
	f := &testFlow{entered: make(chan struct{}), unblock: make(chan struct{})}
	c, _ := s.Start(ctx, "alice", f)
	result := make(chan error, 1)
	go func() { _, err := s.Submit(ctx, "alice", c.ID, "right"); result <- err }()
	<-f.entered
	if _, err := s.Submit(ctx, "alice", c.ID, "right"); !errors.Is(err, ErrBusy) {
		t.Fatal("concurrent submission accepted")
	}
	if err := s.Cancel("alice", c.ID); err != nil {
		t.Fatal(err)
	}
	close(f.unblock)
	if err := <-result; !errors.Is(err, ErrSession) || f.replays.Load() != 0 {
		t.Fatal("cancelled session resumed")
	}
}

func TestAmbiguousSubmissionClosesWithoutReplay(t *testing.T) {
	s := NewSessions()
	defer s.Close()
	ctx := context.Background()
	f := &testFlow{fail: errors.New("network interrupted")}
	c, _ := s.Start(ctx, "alice", f)
	if _, err := s.Submit(ctx, "alice", c.ID, "right"); err == nil {
		t.Fatal("lost error")
	}
	if _, err := s.Refresh(ctx, "alice", c.ID); !errors.Is(err, ErrSession) || f.replays.Load() != 0 {
		t.Fatal("ambiguous request reused")
	}
}
