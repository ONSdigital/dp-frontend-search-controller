package cache

import "sync"

const (
	CensusTopicTitle = "Census"
)

type Topic struct {
	sync.RWMutex
	ID              string
	LocaliseKeyName string
	SubtopicsIDs    []string
}

func (t *Topic) appendSubtopicID(id string) {
	t.Lock()
	defer t.Unlock()

	t.SubtopicsIDs = append(t.SubtopicsIDs, id)
}
