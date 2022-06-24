package cache

import (
	"context"
	"errors"
	"sync"
	"testing"

	topicCliErr "github.com/ONSdigital/dp-topic-api/apierrors"
	"github.com/ONSdigital/dp-topic-api/models"
	"github.com/ONSdigital/dp-topic-api/sdk"
	mockTopic "github.com/ONSdigital/dp-topic-api/sdk/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testCensusTopicID       = "1234"
	testCensusTitle         = "Census"
	testCensusSubTopicID1   = "5678"
	testCensusSubTopicID2   = "1235"
	testCensusSubSubTopicID = "8901"
)

var (
	// root topic level (when GetRootTopicsPublic is called)
	testRootTopics = &models.PublicSubtopics{
		Count:       2,
		Offset:      0,
		Limit:       50,
		TotalCount:  2,
		PublicItems: &[]models.Topic{testCensusRootTopic, testEconomyRootTopic},
	}

	testCensusRootTopic = models.Topic{
		ID:          testCensusTopicID,
		Title:       testCensusTitle,
		SubtopicIds: []string{"5678", "1235"},
	}

	testEconomyRootTopic = models.Topic{
		ID:          "1458",
		Title:       "Economy",
		SubtopicIds: []string{},
	}

	// census sub topic level (when GetSubTopicsPublic is called with `testCensusTopicID` - testRootCensusTopic)
	testCensusSubTopics = &models.PublicSubtopics{
		Count:       3,
		Offset:      0,
		Limit:       50,
		TotalCount:  3,
		PublicItems: &[]models.Topic{testCensusSubTopic1, testCensusSubTopic2},
	}

	testCensusSubTopic1 = models.Topic{
		ID:          testCensusSubTopicID1,
		Title:       "Census Sub 1",
		SubtopicIds: []string{"8901"},
	}

	testCensusSubTopic2 = models.Topic{
		ID:          testCensusSubTopicID2,
		Title:       "Census Sub 2",
		SubtopicIds: []string{},
	}

	// census sub-sub topic level (when GetSubTopicsPublic is called with `testCensusSubTopicID1` - testCensusSubTopic1)
	testCensusSubTopic1SubTopics = &models.PublicSubtopics{
		Count:       1,
		Offset:      0,
		Limit:       50,
		TotalCount:  1,
		PublicItems: &[]models.Topic{testCensusSubTopic1Sub},
	}

	testCensusSubTopic1Sub = models.Topic{
		ID:          testCensusSubSubTopicID,
		Title:       "Census Sub 1 - Sub",
		SubtopicIds: []string{},
	}

	expectedCensusTopicCache = &Topic{
		ID:              testCensusTopicID,
		LocaliseKeyName: testCensusTitle,
		SubtopicsIDs:    []string{testCensusSubTopicID1, testCensusSubTopicID2, testCensusSubSubTopicID},
	}
)

func mockGetSubtopicsIDsPublic(ctx context.Context, subtopicsIDChan chan string, topicClient sdk.Clienter, topLevelTopicID string) (subtopicIDSlice []string) {
	var wg sync.WaitGroup
	wg.Add(2)

	receiveTopic := &Topic{
		ID:              "0000",
		LocaliseKeyName: "test chan receiver",
		SubtopicsIDs:    []string{},
	}

	go func() {
		defer wg.Done()
		getSubtopicsIDsPublic(ctx, subtopicsIDChan, topicClient, topLevelTopicID)
		close(subtopicsIDChan)
	}()

	go func() {
		defer wg.Done()
		for subtopicID := range subtopicsIDChan {
			receiveTopic.appendSubtopicID(subtopicID)
		}
	}()

	wg.Wait()

	return receiveTopic.SubtopicsIDs
}

func TestUpdateCensusTopicPublic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockedTopicClient := &mockTopic.ClienterMock{
		GetRootTopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PublicSubtopics, error) {
			return testRootTopics, nil
		},

		GetSubtopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PublicSubtopics, error) {
			switch id {
			case testCensusTopicID:
				return testCensusSubTopics, nil
			case testCensusSubTopicID1:
				return testCensusSubTopic1SubTopics, nil
			default:
				return nil, errors.New("unexpected error")
			}
		},
	}

	Convey("Given census root topic does exist and has subtopics", t, func() {
		Convey("When UpdateCensusTopicPublic is called", func() {
			respCensusTopicCache, err := UpdateCensusTopicPublic(ctx, mockedTopicClient)()

			Convey("Then the census topic cache is returned", func() {
				So(respCensusTopicCache, ShouldHaveSameTypeAs, expectedCensusTopicCache)
				So(respCensusTopicCache, ShouldResemble, expectedCensusTopicCache)

				Convey("And no error should be returned", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})

	Convey("Given an error in getting root topics from topic-api", t, func() {
		failedRootTopicClient := &mockTopic.ClienterMock{
			GetRootTopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PublicSubtopics, error) {
				return nil, errors.New("unexpected error")
			},
		}

		Convey("When UpdateCensusTopicPublic is called", func() {
			respCensusTopicCache, err := UpdateCensusTopicPublic(ctx, failedRootTopicClient)()

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And the census topic cache returned is nil", func() {
					So(respCensusTopicCache, ShouldBeNil)
				})
			})
		})
	})

	Convey("Given root topics public items is nil", t, func() {
		rootTopicsNilClient := &mockTopic.ClienterMock{
			GetRootTopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PublicSubtopics, error) {
				rootTopicPublicItemsNil := *testRootTopics
				rootTopicPublicItemsNil.PublicItems = nil
				return &rootTopicPublicItemsNil, nil
			},
		}

		Convey("When UpdateCensusTopicPublic is called", func() {
			respCensusTopicCache, err := UpdateCensusTopicPublic(ctx, rootTopicsNilClient)()

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "root topic public items is nil")

				Convey("And the census topic cache returned is nil", func() {
					So(respCensusTopicCache, ShouldBeNil)
				})
			})
		})
	})

	Convey("Given census root topic does not exist", t, func() {
		NonCensusRootTopics := &models.PublicSubtopics{
			Count:       1,
			Offset:      0,
			Limit:       50,
			TotalCount:  1,
			PublicItems: &[]models.Topic{testEconomyRootTopic},
		}

		censusTopicNotExistClient := &mockTopic.ClienterMock{
			GetRootTopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PublicSubtopics, error) {
				return NonCensusRootTopics, nil
			},
		}

		Convey("When UpdateCensusTopicPublic is called", func() {
			respCensusTopicCache, err := UpdateCensusTopicPublic(ctx, censusTopicNotExistClient)()

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "census root topic not found")

				Convey("And the census topic cache returned is nil", func() {
					So(respCensusTopicCache, ShouldBeNil)
				})
			})
		})
	})
}

func TestGetRootTopicCachePublic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	subtopicsIDChan := make(chan string)

	mockedTopicClient := &mockTopic.ClienterMock{
		GetSubtopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PublicSubtopics, error) {
			switch id {
			case testCensusTopicID:
				return testCensusSubTopics, nil
			case testCensusSubTopicID1:
				return testCensusSubTopic1SubTopics, nil
			default:
				return nil, errors.New("unexpected error")
			}
		},
	}

	Convey("Given topic has subtopics", t, func() {

		Convey("When getRootTopicCachePublic is called", func() {
			respCensusTopicCache := getRootTopicCachePublic(ctx, subtopicsIDChan, mockedTopicClient, testCensusRootTopic)

			Convey("Then the census topic cache is returned", func() {
				So(respCensusTopicCache, ShouldResemble, expectedCensusTopicCache)
			})
		})
	})
}

func TestGetSubtopicsIDsPublic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockedTopicClient := &mockTopic.ClienterMock{
		GetSubtopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PublicSubtopics, error) {
			switch id {
			case testCensusTopicID:
				return testCensusSubTopics, nil
			case testCensusSubTopicID1:
				return testCensusSubTopic1SubTopics, nil
			case testCensusSubTopicID2:
				return nil, topicCliErr.ErrNotFound
			default:
				return nil, errors.New("unexpected error")
			}
		},
	}

	Convey("Given topic has subtopics", t, func() {
		subtopicsIDChan := make(chan string)

		Convey("When getSubtopicsIDsPublic is called", func() {
			subTopicsIDSlice := mockGetSubtopicsIDsPublic(ctx, subtopicsIDChan, mockedTopicClient, testCensusTopicID)

			Convey("Then subtopic ids should be sent to subtopicsIDChan channel", func() {
				So(subTopicsIDSlice, ShouldHaveLength, 3)
				So(subTopicsIDSlice, ShouldResemble, []string{testCensusSubTopicID1, testCensusSubTopicID2, testCensusSubSubTopicID})
			})
		})
	})

	Convey("Given topic has no subtopics", t, func() {
		subtopicsIDChan := make(chan string)

		Convey("When getSubtopicsIDsPublic is called", func() {
			subTopicsIDSlice := mockGetSubtopicsIDsPublic(ctx, subtopicsIDChan, mockedTopicClient, testCensusSubTopicID2)

			Convey("Then no subtopic ids should be sent to subtopicsIDChan channel", func() {
				So(subTopicsIDSlice, ShouldHaveLength, 0)
			})
		})
	})

	Convey("Given an error in getting sub topics from topic-api", t, func() {
		subtopicsIDChan := make(chan string)

		failedGetSubtopicClient := &mockTopic.ClienterMock{
			GetSubtopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PublicSubtopics, error) {
				return nil, errors.New("unexpected error")
			},
		}

		Convey("When getSubtopicsIDsPublic is called", func() {
			subTopicsIDSlice := mockGetSubtopicsIDsPublic(ctx, subtopicsIDChan, failedGetSubtopicClient, testCensusTopicID)

			Convey("Then no subtopic ids should be sent to subtopicsIDChan channel", func() {
				So(subTopicsIDSlice, ShouldHaveLength, 0)
			})
		})
	})

	Convey("Given sub topics public items is nil", t, func() {
		subtopicsIDChan := make(chan string)

		subtopicItemsNilClient := &mockTopic.ClienterMock{
			GetSubtopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PublicSubtopics, error) {
				topicItemsNil := *testCensusSubTopics
				topicItemsNil.PublicItems = nil
				return &topicItemsNil, nil
			},
		}

		Convey("When getSubtopicsIDsPublic is called", func() {
			subTopicsIDSlice := mockGetSubtopicsIDsPublic(ctx, subtopicsIDChan, subtopicItemsNilClient, testCensusTopicID)

			Convey("Then no subtopic ids should be sent to subtopicsIDChan channel", func() {
				So(subTopicsIDSlice, ShouldHaveLength, 0)
			})
		})
	})
}
