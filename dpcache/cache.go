package dpcache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ONSdigital/log.go/v2/log"
)

// Cacher defines the required methods to initialise a cache
type Cacher interface {
	Close()
	Get(key string) (interface{}, bool)
	Set(key string, data interface{})
	AddUpdateFunc(key string, updateFunc func() (interface{}, error))
	StartUpdates(ctx context.Context, channel chan error)
}

// Cache contains all the information to start, update and close caching data
type Cache struct {
	data           sync.Map
	updateInterval *time.Duration
	close          chan struct{}
	UpdateFuncs    map[string]func() (interface{}, error)
}

// NewCache create a cache object which will update at every updateInterval
// If updateInterval is nil, this means that the cache will only be updated once at the start of a service
func NewCache(ctx context.Context, updateInterval *time.Duration) (*Cache, error) {
	if updateInterval != nil {
		if *updateInterval <= 0 {
			err := errors.New("cache update interval duration is less than or equal to 0")
			log.Error(ctx, "invalid cache update interval given", err)
			return nil, err
		}
	}

	return &Cache{
		data:           sync.Map{},
		updateInterval: updateInterval,
		close:          make(chan struct{}),
		UpdateFuncs:    make(map[string]func() (interface{}, error)),
	}, nil
}

// Get retrieves the specific value for the specified key stored in `data` within the `Cache`
func (dc *Cache) Get(key string) (interface{}, bool) {
	return dc.data.Load(key)
}

// Set adds the specified value with the specified key in `data` within the `Cache`
func (dc *Cache) Set(key string, data interface{}) {
	dc.data.Store(key, data)
}

// Close closes the caching of data when called where the data will no longer be updated and the data itself is reset
func (dc *Cache) Close() {
	if dc.updateInterval != nil {
		dc.close <- struct{}{}
		for key := range dc.UpdateFuncs {
			dc.data.Store(key, "")
		}
		dc.UpdateFuncs = make(map[string]func() (interface{}, error))
	}
}

// AddUpdateFunc adds an update function to the cache for a specific data corresponding to the `key` passed to the function
// This update function will then be triggered once or at every fixed interval as per the prior setup of the TopicCache
func (dc *Cache) AddUpdateFunc(key string, updateFunc func() (interface{}, error)) {
	dc.UpdateFuncs[key] = updateFunc
}

// UpdateContent calls all the update functions with a key value stored in the Cache to update the relevant data with the same key values
func (dc *Cache) UpdateContent(ctx context.Context) error {
	for key, updateFunc := range dc.UpdateFuncs {
		updatedContent, err := updateFunc()
		if err != nil {
			return fmt.Errorf("failed to update search cache for %s. error: %v", key, err)
		}
		dc.Set(key, updatedContent)
	}
	return nil
}

// StartUpdates informs the cache to start updating the cache data once called and then at every update interval which was configured when setting up the cache
func (dc *Cache) StartUpdates(ctx context.Context, errorChannel chan error) {
	if len(dc.UpdateFuncs) == 0 {
		return
	}

	err := dc.UpdateContent(ctx)
	if err != nil {
		errorChannel <- err
		dc.Close()
		return
	}

	if dc.updateInterval != nil {
		ticker := time.NewTicker(*dc.updateInterval)

		for {
			select {
			case <-ticker.C:
				err := dc.UpdateContent(ctx)
				if err != nil {
					log.Error(ctx, err.Error(), err)
				}

			case <-dc.close:
				return
			case <-ctx.Done():
				return
			}
		}
	}
}
