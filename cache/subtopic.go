package cache

import (
	"strings"
	"sync"
	"time"
)

// Subtopics contains a list of subtopics in map form with addition to mutex locking
// The subtopicsMap is used to keep a record of subtopics to be later used to generate the subtopics id `query` for a topic
// and to check if the subtopic id given by a user exists
type Subtopics struct {
	mutex        *sync.RWMutex
	subtopicsMap map[string]Subtopic
}

// Subtopic represents the data which is cached for a subtopic to be used by the dp-frontend-search-controller
type Subtopic struct {
	ID              string
	LocaliseKeyName string
	Slug            string
	ReleaseDate     *time.Time
	ParentID        string
	ParentSlug      string
}

// GetNewSubTopicsMap creates a new subtopics id map to store subtopic ids with addition to mutex locking
func NewSubTopicsMap() *Subtopics {
	return &Subtopics{
		mutex:        &sync.RWMutex{},
		subtopicsMap: make(map[string]Subtopic),
	}
}

// Get returns subtopic for given key (id)
func (t *Subtopics) Get(key string) (Subtopic, bool) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	subtopic, exists := t.subtopicsMap[key]
	return subtopic, exists
}

// GetBySlugAndParentSlug returns a subtopic that matches the given topic slug and parentSlug
func (t *Subtopics) GetBySlugAndParentSlug(slug, parentSlug string) (Subtopic, bool) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	for _, subtopic := range t.subtopicsMap {
		if subtopic.Slug == slug && subtopic.ParentSlug == parentSlug {
			return subtopic, true
		}
	}
	return Subtopic{}, false
}

// GetSubtopics returns an array of subtopics
func (t *Subtopics) GetSubtopics() (subtopics []Subtopic) {
	if t.subtopicsMap == nil {
		return
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	for _, subtopic := range t.subtopicsMap {
		subtopics = append(subtopics, subtopic)
	}

	return subtopics
}

// CheckTopicIDExists returns subtopic for given key (id)
func (t *Subtopics) CheckTopicIDExists(key string) bool {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if _, ok := t.subtopicsMap[key]; !ok {
		return false
	}

	return true
}

// GetSubtopicsIDsQuery gets the subtopics ID query for a topic
func (t *Subtopics) GetSubtopicsIDsQuery() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	ids := make([]string, 0, len(t.subtopicsMap))

	for id := range t.subtopicsMap {
		ids = append(ids, id)
	}

	return strings.Join(ids, ",")
}

// AppendSubtopicID appends the subtopic id to the map stored in SubtopicsIDs with consideration to mutex locking
func (t *Subtopics) AppendSubtopicID(id string, subtopic Subtopic) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.subtopicsMap == nil {
		t.subtopicsMap = make(map[string]Subtopic)
	}

	t.subtopicsMap[id] = subtopic
}
