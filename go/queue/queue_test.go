package queue

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func fixedClock() func() time.Time {
	t := time.Date(2026, 10, 8, 20, 0, 0, 0, time.UTC)
	return func() time.Time { return t }
}

func TestDuplicateSongIsRejected(t *testing.T) {
	s := NewStore(fixedClock())
	if _, err := s.Add("bar", "song-1", "A", "t1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Add("bar", "song-1", "A", "t2"); !errors.Is(err, ErrDuplicateSong) {
		t.Fatalf("want ErrDuplicateSong, got %v", err)
	}
}

func TestTableLimit(t *testing.T) {
	s := NewStore(fixedClock())
	s.Add("bar", "s1", "A", "t1")
	s.Add("bar", "s2", "B", "t1")
	if _, err := s.Add("bar", "s3", "C", "t1"); !errors.Is(err, ErrTableLimit) {
		t.Fatalf("want ErrTableLimit, got %v", err)
	}
}

func TestVotesReorderAndTiesKeepArrivalOrder(t *testing.T) {
	s := NewStore(fixedClock())
	a, _ := s.Add("bar", "s1", "A", "t1")
	b, _ := s.Add("bar", "s2", "B", "t2")
	c, _ := s.Add("bar", "s3", "C", "t3")

	s.Vote("bar", c.ID, "t9")

	got := s.List("bar")
	want := []string{c.ID, a.ID, b.ID} // C tiene un voto; A y B empatan y A llegó antes
	for i, id := range want {
		if got[i].ID != id {
			t.Fatalf("position %d: want %s, got %s", i, id, got[i].ID)
		}
	}
}

func TestOneVotePerTable(t *testing.T) {
	s := NewStore(fixedClock())
	it, _ := s.Add("bar", "s1", "A", "t1")
	if _, err := s.Vote("bar", it.ID, "t2"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Vote("bar", it.ID, "t2"); !errors.Is(err, ErrAlreadyVoted) {
		t.Fatalf("want ErrAlreadyVoted, got %v", err)
	}
}

func TestAdvancePopsHeadAndFreesTableSlot(t *testing.T) {
	s := NewStore(fixedClock())
	a, _ := s.Add("bar", "s1", "A", "t1")
	s.Add("bar", "s2", "B", "t1")

	playing, err := s.Advance("bar")
	if err != nil || playing.ID != a.ID {
		t.Fatalf("want %s playing, got %v / %v", a.ID, playing.ID, err)
	}
	// La mesa t1 recupera un cupo al sonar su canción.
	if _, err := s.Add("bar", "s3", "C", "t1"); err != nil {
		t.Fatalf("slot should be free: %v", err)
	}
}

func TestAdvanceEmpty(t *testing.T) {
	s := NewStore(fixedClock())
	if _, err := s.Advance("bar"); !errors.Is(err, ErrEmpty) {
		t.Fatalf("want ErrEmpty, got %v", err)
	}
}

func TestVenuesAreIsolated(t *testing.T) {
	s := NewStore(fixedClock())
	s.Add("bar-a", "s1", "A", "t1")
	if _, err := s.Add("bar-b", "s1", "A", "t1"); err != nil {
		t.Fatalf("same song in another venue must be allowed: %v", err)
	}
}

func TestConcurrentAddsRespectDuplicateRule(t *testing.T) {
	s := NewStore(nil)
	var wg sync.WaitGroup
	var mu sync.Mutex
	ok := 0
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// 50 mesas distintas piden la misma canción a la vez.
			if _, err := s.Add("bar", "same-song", "X", "t"+string(rune('a'+i%26))+string(rune('0'+i/26))); err == nil {
				mu.Lock()
				ok++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	if ok != 1 {
		t.Fatalf("exactly one add should win, got %d", ok)
	}
}
