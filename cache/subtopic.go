package cache

import (
	"sort"
	"strings"
	"sync"
)

// SubtopicsIDs contains a list of subtopics in map form with addition to mutex locking
// The subtopicsMap is used to keep a record of subtopics to be later used to generate the subtopics id `query` for a topic
// and to check if the subtopic id given by a user exists
type SubtopicsIDs struct {
	sync.RWMutex
	subtopicsMap map[string]bool
}

// GetNewSubTopicsMap creates a new subtopics id map to store subtopic ids with addition to mutex locking
func NewSubTopicsMap() *SubtopicsIDs {
	return &SubtopicsIDs{
		subtopicsMap: make(map[string]bool),
	}
}

// Get returns a bool value for the given key (id) to inform that the subtopic id exists
func (t *SubtopicsIDs) Get(key string) bool {
	t.RLock()
	defer t.RUnlock()

	return t.subtopicsMap[key]
}

// GetSubtopicsIDsQuery gets the subtopics ID query for a topic
func (t *SubtopicsIDs) GetSubtopicsIDsQuery() string {
	t.RLock()
	defer t.RUnlock()

	ids := make([]string, 0, len(t.subtopicsMap))

	for id := range t.subtopicsMap {
		ids = append(ids, id)
	}

	sort.Strings(ids)

	return strings.Join(ids, ",")
}

// AppendSubtopicID appends the subtopic id to the map stored in SubtopicsIDs with consideration to mutex locking
func (t *SubtopicsIDs) AppendSubtopicID(id string) {
	t.Lock()
	defer t.Unlock()

	if t.subtopicsMap == nil {
		t.subtopicsMap = make(map[string]bool)
	}

	t.subtopicsMap[id] = true
}
