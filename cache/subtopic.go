package cache

import (
	"strings"
	"sync"
)

// SubtopicsIDs contains a list of subtopics in map form with addition to mutex locking
// The subtopicsMap is used to keep a record of subtopics to be later used to generate the subtopics id `query` for a topic
// and to check if the subtopic id given by a user exists
type SubtopicsIDs struct {
	mutex        *sync.RWMutex
	subtopicsMap map[string]Subtopic
}

type Subtopic struct {
	LocaliseKeyName string
	ReleaseDate     string
}

// GetNewSubTopicsMap creates a new subtopics id map to store subtopic ids with addition to mutex locking
func NewSubTopicsMap() *SubtopicsIDs {
	return &SubtopicsIDs{
		mutex:        &sync.RWMutex{},
		subtopicsMap: make(map[string]Subtopic),
	}
}

// Get returns subtopic for given key (id)
func (t *SubtopicsIDs) Get(key string) Subtopic {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.subtopicsMap[key]
}

// CheckTopicIDExsits returns subtopic for given key (id)
func (t *SubtopicsIDs) CheckTopicIDExsits(key string) bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if _, ok := t.subtopicsMap[key]; !ok {
		return false
	}

	return true
}

// GetSubtopicsIDsQuery gets the subtopics ID query for a topic
func (t *SubtopicsIDs) GetSubtopicsIDsQuery() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	ids := make([]string, 0, len(t.subtopicsMap))

	for id := range t.subtopicsMap {
		ids = append(ids, id)
	}

	return strings.Join(ids, ",")
}

// AppendSubtopicID appends the subtopic id to the map stored in SubtopicsIDs with consideration to mutex locking
func (t *SubtopicsIDs) AppendSubtopicID(id string, subtopic Subtopic) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.subtopicsMap == nil {
		t.subtopicsMap = make(map[string]Subtopic)
	}

	t.subtopicsMap[id] = subtopic
}
