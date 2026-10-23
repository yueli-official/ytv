package verification

import (
	"context"
	"crypto/rand"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/yueli-official/ytv/api/internal/source"
)

var (
	ErrSession = errors.New("verification session missing or expired")
	ErrBusy    = errors.New("verification session is busy")
	ErrLimit   = errors.New("verification session limit reached")
	ErrAnswer  = errors.New("verification answer is invalid")
)

type CaptchaImage struct {
	MIME  string
	Bytes []byte
}
type Challenge struct {
	ID        string
	ExpiresAt time.Time
	Image     CaptchaImage
}
type Submission struct {
	Accepted bool
	Page     source.Page
}

// Flow owns one isolated upstream session and immutable original operation.
// Submit must report acceptance from an explicit upstream result, not HTTP 200.
// Close must be nonblocking, safe concurrently with network work, and cancel
// owned requests. Other Flow methods run serially within a session.
type Flow interface {
	Refresh(context.Context) (CaptchaImage, error)
	Submit(context.Context, string) (bool, error)
	Resume(context.Context) (source.Page, error)
	Close()
}

type session struct {
	owner     string
	expires   time.Time
	flow      Flow
	mu        sync.Mutex
	consumed  bool
	closeOnce sync.Once
}

func (s *session) close() { s.closeOnce.Do(s.flow.Close) }

// Sessions has no sweeper goroutine: expiry is bounded and checked on every
// operation. Close belongs to the runtime owner. IDs never grant ownership.
type Sessions struct {
	mu       sync.Mutex
	entries  map[string]*session
	clock    func() time.Time
	ttl      time.Duration
	limit    int
	perOwner int
	closed   bool
}

func NewSessions() *Sessions {
	return &Sessions{entries: map[string]*session{}, clock: time.Now, ttl: 3 * time.Minute, limit: 128, perOwner: 2}
}

// Start takes ownership of flow, including when capacity or validation fails.
func (s *Sessions) Start(ctx context.Context, owner string, flow Flow) (Challenge, error) {
	if flow == nil {
		return Challenge{}, ErrSession
	}
	if strings.TrimSpace(owner) == "" {
		flow.Close()
		return Challenge{}, ErrSession
	}
	entry := &session{owner: owner, expires: s.clock().Add(s.ttl), flow: flow}
	entry.mu.Lock()
	defer entry.mu.Unlock()
	s.mu.Lock()
	s.pruneLocked()
	count := 0
	for _, item := range s.entries {
		if item.owner == owner {
			count++
		}
	}
	if s.closed || len(s.entries) >= s.limit || count >= s.perOwner {
		s.mu.Unlock()
		flow.Close()
		return Challenge{}, ErrLimit
	}
	id := rand.Text()
	s.entries[id] = entry
	s.mu.Unlock()
	image, err := flow.Refresh(ctx)
	if err != nil {
		s.remove(id, entry)
		return Challenge{}, err
	}
	if !s.current(id, entry) {
		return Challenge{}, ErrSession
	}
	return Challenge{ID: id, ExpiresAt: entry.expires, Image: image}, nil
}

func (s *Sessions) Refresh(ctx context.Context, owner, id string) (Challenge, error) {
	entry, err := s.acquire(owner, id)
	if err != nil {
		return Challenge{}, err
	}
	defer entry.mu.Unlock()
	image, err := entry.flow.Refresh(ctx)
	if err != nil {
		s.remove(id, entry)
		return Challenge{}, err
	}
	if !s.current(id, entry) {
		return Challenge{}, ErrSession
	}
	return Challenge{ID: id, ExpiresAt: entry.expires, Image: image}, nil
}

func (s *Sessions) Submit(ctx context.Context, owner, id, answer string) (Submission, error) {
	answer = strings.TrimSpace(answer)
	if len(answer) == 0 || len(answer) > 32 {
		return Submission{}, ErrAnswer
	}
	entry, err := s.acquire(owner, id)
	if err != nil {
		return Submission{}, err
	}
	defer entry.mu.Unlock()
	accepted, err := entry.flow.Submit(ctx, answer)
	if err != nil {
		s.remove(id, entry)
		return Submission{}, err
	}
	if !s.current(id, entry) {
		return Submission{}, ErrSession
	}
	if !accepted {
		return Submission{}, nil
	}
	// Consume before replay: no retry can replay the original operation again.
	s.mu.Lock()
	if s.entries[id] != entry {
		s.mu.Unlock()
		return Submission{}, ErrSession
	}
	entry.consumed = true
	s.mu.Unlock()
	defer s.remove(id, entry)
	page, err := entry.flow.Resume(ctx)
	if !s.current(id, entry) {
		return Submission{}, ErrSession
	}
	return Submission{Accepted: true, Page: page}, err
}

func (s *Sessions) Cancel(owner, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked()
	entry := s.entries[id]
	if entry == nil || entry.owner != owner {
		return ErrSession
	}
	delete(s.entries, id)
	entry.close()
	return nil
}

func (s *Sessions) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	for id, entry := range s.entries {
		delete(s.entries, id)
		entry.close()
	}
}
func (s *Sessions) pruneLocked() {
	for id, entry := range s.entries {
		if !s.clock().Before(entry.expires) {
			delete(s.entries, id)
			entry.close()
		}
	}
}
func (s *Sessions) acquire(owner, id string) (*session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked()
	entry := s.entries[id]
	if entry == nil || entry.owner != owner || entry.consumed {
		return nil, ErrSession
	}
	if !entry.mu.TryLock() {
		return nil, ErrBusy
	}
	return entry, nil
}
func (s *Sessions) current(id string, entry *session) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pruneLocked()
	return s.entries[id] == entry
}
func (s *Sessions) remove(id string, entry *session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.entries[id] == entry {
		delete(s.entries, id)
	}
	entry.close()
}
