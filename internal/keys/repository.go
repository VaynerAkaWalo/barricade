package keys

import (
	"context"
	"sort"
	"sync"
)

type InMemoryRepository struct {
	mu   sync.RWMutex
	keys map[KeyId]*Key
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		keys: make(map[KeyId]*Key),
	}
}

func (r *InMemoryRepository) Save(ctx context.Context, key *Key) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	storedKey := *key
	r.keys[key.Id] = &storedKey
	return nil
}

func (r *InMemoryRepository) FindById(ctx context.Context, id KeyId) (*Key, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key, exists := r.keys[id]
	if !exists {
		return nil, ErrKeyNotFound
	}
	return key, nil
}

func (r *InMemoryRepository) FindLatest(ctx context.Context, algorithm Algorithm) (*Key, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var candidates []*Key
	for _, key := range r.keys {
		if key.Algorithm == algorithm {
			candidates = append(candidates, key)
		}
	}

	if len(candidates) == 0 {
		return nil, ErrNoSigningKey
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].CreatedAt > candidates[j].CreatedAt
	})

	return candidates[0], nil
}

func (r *InMemoryRepository) FindAll(ctx context.Context) ([]*Key, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Key, 0, len(r.keys))
	for _, key := range r.keys {
		result = append(result, key)
	}

	return result, nil
}
