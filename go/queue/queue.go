// Package queue tiene la lógica de la cola de un local, sin nada de HTTP.
package queue

import (
	"errors"
	"sort"
	"strconv"
	"sync"
	"time"
)

// MaxPendingPerTable es cuántos pedidos pendientes puede tener una mesa.
const MaxPendingPerTable = 2

var (
	ErrDuplicateSong = errors.New("song already in queue")
	ErrTableLimit    = errors.New("table has reached its pending request limit")
	ErrAlreadyVoted  = errors.New("table already voted for this item")
	ErrNotFound      = errors.New("item not found")
	ErrEmpty         = errors.New("queue is empty")
)

// Item es una canción pedida.
type Item struct {
	ID        string    `json:"id"`
	SongID    string    `json:"songId"`
	Title     string    `json:"title"`
	TableID   string    `json:"tableId"`
	Votes     int       `json:"votes"`
	CreatedAt time.Time `json:"createdAt"`

	voters map[string]bool
	seq    int64 // desempata dos ítems creados en el mismo instante
}

type venueQueue struct {
	items []*Item
}

// Store guarda una cola por local. Es seguro para uso concurrente.
type Store struct {
	mu     sync.Mutex
	venues map[string]*venueQueue
	seq    int64
	now    func() time.Time
}

// NewStore crea un Store vacío. now permite fijar el reloj en pruebas.
func NewStore(now func() time.Time) *Store {
	if now == nil {
		now = time.Now
	}
	return &Store{venues: map[string]*venueQueue{}, now: now}
}

func (s *Store) venue(id string) *venueQueue {
	v, ok := s.venues[id]
	if !ok {
		v = &venueQueue{}
		s.venues[id] = v
	}
	return v
}

// Add pide una canción para una mesa.
func (s *Store) Add(venueID, songID, title, tableID string) (Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v := s.venue(venueID)
	pending := 0
	for _, it := range v.items {
		if it.SongID == songID {
			return Item{}, ErrDuplicateSong
		}
		if it.TableID == tableID {
			pending++
		}
	}
	if pending >= MaxPendingPerTable {
		return Item{}, ErrTableLimit
	}

	s.seq++
	it := &Item{
		ID:        "itm_" + strconv.FormatInt(s.seq, 10),
		SongID:    songID,
		Title:     title,
		TableID:   tableID,
		CreatedAt: s.now(),
		voters:    map[string]bool{},
		seq:       s.seq,
	}
	v.items = append(v.items, it)
	return *it, nil
}

// Vote suma el voto de una mesa a un ítem.
func (s *Store) Vote(venueID, itemID, tableID string) (Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, it := range s.venue(venueID).items {
		if it.ID != itemID {
			continue
		}
		if it.voters[tableID] {
			return Item{}, ErrAlreadyVoted
		}
		it.voters[tableID] = true
		it.Votes++
		return *it, nil
	}
	return Item{}, ErrNotFound
}

// List devuelve la cola ordenada: más votos primero, luego el más antiguo.
func (s *Store) List(venueID string) []Item {
	s.mu.Lock()
	defer s.mu.Unlock()

	sorted := s.sorted(venueID)
	out := make([]Item, len(sorted))
	for i, it := range sorted {
		out[i] = *it
	}
	return out
}

// Advance saca el primero de la cola y lo devuelve como "sonando ahora".
func (s *Store) Advance(venueID string) (Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sorted := s.sorted(venueID)
	if len(sorted) == 0 {
		return Item{}, ErrEmpty
	}
	head := sorted[0]
	v := s.venue(venueID)
	for i, it := range v.items {
		if it == head {
			v.items = append(v.items[:i], v.items[i+1:]...)
			break
		}
	}
	return *head, nil
}

// sorted asume que s.mu ya está tomado.
func (s *Store) sorted(venueID string) []*Item {
	items := append([]*Item(nil), s.venue(venueID).items...)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Votes != items[j].Votes {
			return items[i].Votes > items[j].Votes
		}
		return items[i].seq < items[j].seq
	})
	return items
}
