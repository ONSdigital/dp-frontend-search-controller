package private

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ONSdigital/dp-frontend-search-controller/cache"
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
	// root topic level (when GetRootTopics is called)
	testRootTopicsPrivate = &models.PrivateSubtopics{
		Count:        2,
		Offset:       0,
		Limit:        50,
		TotalCount:   2,
		PrivateItems: &[]models.TopicResponse{testCensusRootTopicPrivate, testEconomyRootTopicPrivate},
	}

	testCensusRootTopicPrivate = models.TopicResponse{
		ID:      testCensusTopicID,
		Next:    &testCensusRootTopic,
		Current: &testCensusRootTopic,
	}

	testEconomyRootTopicPrivate = models.TopicResponse{
		ID:      "1458",
		Next:    &testEconomyRootTopic,
		Current: &testEconomyRootTopic,
	}

	// census sub topic level (when GetSubTopics is called with `testCensusTopicID` - testRootCensusTopic)
	testCensusSubTopicsPrivate = &models.PrivateSubtopics{
		Count:        3,
		Offset:       0,
		Limit:        50,
		TotalCount:   3,
		PrivateItems: &[]models.TopicResponse{testCensusSubTopic1Private, testCensusSubTopic2Private},
	}

	testCensusSubTopic1Private = models.TopicResponse{
		ID:      testCensusSubTopicID1,
		Next:    &testCensusSubTopic1,
		Current: &testCensusSubTopic1,
	}

	testCensusSubTopic2Private = models.TopicResponse{
		ID:      testCensusSubTopicID2,
		Next:    &testCensusSubTopic2,
		Current: &testCensusSubTopic2,
	}

	// census sub-sub topic level (when GetSubTopics is called with `testCensusSubTopicID1` - testCensusSubTopic1)
	testCensusSubTopic1SubTopicsPrivate = &models.PrivateSubtopics{
		Count:        1,
		Offset:       0,
		Limit:        50,
		TotalCount:   1,
		PrivateItems: &[]models.TopicResponse{testCensusSubTopic1SubPrivate},
	}

	testCensusSubTopic1SubPrivate = models.TopicResponse{
		ID:      testCensusSubSubTopicID,
		Next:    &testCensusSubTopic1Sub,
		Current: &testCensusSubTopic1Sub,
	}
)

var (
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

	testCensusSubTopic1Sub = models.Topic{
		ID:          testCensusSubSubTopicID,
		Title:       "Census Sub 1 - Sub",
		SubtopicIds: []string{},
	}

	expectedCensusTopicCache = &cache.Topic{
		ID:               testCensusTopicID,
		LocaliseKeyName:  testCensusTitle,
		SubtopicsIDQuery: fmt.Sprintf("%s,%s,%s", testCensusSubTopicID1, testCensusSubTopicID2, testCensusSubSubTopicID),
	}
)

func mockGetSubtopicsIDsPrivate(ctx context.Context, subtopicsIDChan chan string, topicClient sdk.Clienter, topLevelTopicID string) string {
	var rootTopic models.Topic

	switch topLevelTopicID {
	case testCensusSubTopicID2:
		rootTopic = testCensusSubTopic2
	default:
		rootTopic = testCensusRootTopic
	}

	testTopicCache := getRootTopicCachePrivate(ctx, "", subtopicsIDChan, topicClient, rootTopic)

	return testTopicCache.SubtopicsIDQuery
}

func TestUpdateCensusTopic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockedTopicClient := &mockTopic.ClienterMock{
		GetRootTopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PrivateSubtopics, error) {
			return testRootTopicsPrivate, nil
		},

		GetSubtopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PrivateSubtopics, error) {
			switch id {
			case testCensusTopicID:
				return testCensusSubTopicsPrivate, nil
			case testCensusSubTopicID1:
				return testCensusSubTopic1SubTopicsPrivate, nil
			default:
				return nil, errors.New("unexpected error")
			}
		},
	}

	Convey("Given census root topic does exist and has subtopics", t, func() {
		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache, err := UpdateCensusTopic(ctx, "", mockedTopicClient)()

			Convey("Then the census topic cache is returned", func() {
				So(respCensusTopicCache, ShouldNotBeNil)

				So(respCensusTopicCache.ID, ShouldEqual, expectedCensusTopicCache.ID)
				So(respCensusTopicCache.LocaliseKeyName, ShouldEqual, expectedCensusTopicCache.LocaliseKeyName)

				So(respCensusTopicCache.SubtopicsIDQuery, ShouldContainSubstring, testCensusSubTopicID1)
				So(respCensusTopicCache.SubtopicsIDQuery, ShouldContainSubstring, testCensusSubTopicID2)
				So(respCensusTopicCache.SubtopicsIDQuery, ShouldContainSubstring, testCensusSubSubTopicID)

				Convey("And no error should be returned", func() {
					So(err, ShouldBeNil)
				})
			})
		})
	})

	Convey("Given an error in getting root topics from topic-api", t, func() {
		failedRootTopicClient := &mockTopic.ClienterMock{
			GetRootTopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PrivateSubtopics, error) {
				return nil, errors.New("unexpected error")
			},
		}

		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache, err := UpdateCensusTopic(ctx, "", failedRootTopicClient)()

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And the census topic cache returned is nil", func() {
					So(respCensusTopicCache, ShouldBeNil)
				})
			})
		})
	})

	Convey("Given root topics private items is nil", t, func() {
		rootTopicsNilClient := &mockTopic.ClienterMock{
			GetRootTopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PrivateSubtopics, error) {
				rootTopicPrivateItemsNil := *testRootTopicsPrivate
				rootTopicPrivateItemsNil.PrivateItems = nil
				return &rootTopicPrivateItemsNil, nil
			},
		}

		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache, err := UpdateCensusTopic(ctx, "", rootTopicsNilClient)()

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
		NonCensusRootTopics := &models.PrivateSubtopics{
			Count:        1,
			Offset:       0,
			Limit:        50,
			TotalCount:   1,
			PrivateItems: &[]models.TopicResponse{testEconomyRootTopicPrivate},
		}

		censusTopicNotExistClient := &mockTopic.ClienterMock{
			GetRootTopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PrivateSubtopics, error) {
				return NonCensusRootTopics, nil
			},
		}

		Convey("When UpdateCensusTopicPrivate is called", func() {
			respCensusTopicCache, err := UpdateCensusTopic(ctx, "", censusTopicNotExistClient)()

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

func TestGetRootTopicCachePrivate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	subtopicsIDChan := make(chan string)

	mockedTopicClient := &mockTopic.ClienterMock{
		GetSubtopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PrivateSubtopics, error) {
			switch id {
			case testCensusTopicID:
				return testCensusSubTopicsPrivate, nil
			case testCensusSubTopicID1:
				return testCensusSubTopic1SubTopicsPrivate, nil
			default:
				return nil, errors.New("unexpected error")
			}
		},
	}

	Convey("Given topic has subtopics", t, func() {

		Convey("When getRootTopicCachePrivate is called", func() {
			respCensusTopicCache := getRootTopicCachePrivate(ctx, "", subtopicsIDChan, mockedTopicClient, testCensusRootTopic)

			Convey("Then the census topic cache is returned", func() {
				So(respCensusTopicCache, ShouldNotBeNil)
				So(respCensusTopicCache.ID, ShouldEqual, expectedCensusTopicCache.ID)
				So(respCensusTopicCache.LocaliseKeyName, ShouldEqual, expectedCensusTopicCache.LocaliseKeyName)

				So(respCensusTopicCache.SubtopicsIDQuery, ShouldContainSubstring, testCensusSubTopicID1)
				So(respCensusTopicCache.SubtopicsIDQuery, ShouldContainSubstring, testCensusSubTopicID2)
				So(respCensusTopicCache.SubtopicsIDQuery, ShouldContainSubstring, testCensusSubSubTopicID)
			})
		})
	})
}

func TestGetSubtopicsIDsPrivate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockedTopicClient := &mockTopic.ClienterMock{
		GetSubtopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PrivateSubtopics, error) {
			switch id {
			case testCensusTopicID:
				return testCensusSubTopicsPrivate, nil
			case testCensusSubTopicID1:
				return testCensusSubTopic1SubTopicsPrivate, nil
			case testCensusSubTopicID2:
				return nil, topicCliErr.ErrNotFound
			default:
				return nil, errors.New("unexpected error")
			}
		},
	}

	Convey("Given topic has subtopics", t, func() {
		subtopicsIDChan := make(chan string)

		Convey("When getSubtopicsIDsPrivate is called", func() {
			subTopicsIDQuery := mockGetSubtopicsIDsPrivate(ctx, subtopicsIDChan, mockedTopicClient, testCensusTopicID)

			Convey("Then subtopic ids should be sent to subtopicsIDChan channel", func() {
				So(subTopicsIDQuery, ShouldNotBeEmpty)
				So(subTopicsIDQuery, ShouldContainSubstring, testCensusSubTopicID1)
				So(subTopicsIDQuery, ShouldContainSubstring, testCensusSubTopicID2)
				So(subTopicsIDQuery, ShouldContainSubstring, testCensusSubSubTopicID)
			})
		})
	})

	Convey("Given topic has no subtopics", t, func() {
		subtopicsIDChan := make(chan string)

		Convey("When getSubtopicsIDsPrivate is called", func() {
			subTopicsIDQuery := mockGetSubtopicsIDsPrivate(ctx, subtopicsIDChan, mockedTopicClient, testCensusSubTopicID2)

			Convey("Then no subtopic ids should be sent to subtopicsIDChan channel", func() {
				So(subTopicsIDQuery, ShouldBeEmpty)
			})
		})
	})

	Convey("Given an error in getting sub topics from topic-api", t, func() {
		subtopicsIDChan := make(chan string)

		failedGetSubtopicClient := &mockTopic.ClienterMock{
			GetSubtopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PrivateSubtopics, error) {
				return nil, errors.New("unexpected error")
			},
		}

		Convey("When getSubtopicsIDsPrivate is called", func() {
			subTopicsIDQuery := mockGetSubtopicsIDsPrivate(ctx, subtopicsIDChan, failedGetSubtopicClient, testCensusTopicID)

			Convey("Then no subtopic ids should be sent to subtopicsIDChan channel", func() {
				So(subTopicsIDQuery, ShouldBeEmpty)
			})
		})
	})

	Convey("Given sub topics private items is nil", t, func() {
		subtopicsIDChan := make(chan string)

		subtopicItemsNilClient := &mockTopic.ClienterMock{
			GetSubtopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PrivateSubtopics, error) {
				topicItemsNil := *testCensusSubTopicsPrivate
				topicItemsNil.PrivateItems = nil
				return &topicItemsNil, nil
			},
		}

		Convey("When getSubtopicsIDsPrivate is called", func() {
			subTopicsIDQuery := mockGetSubtopicsIDsPrivate(ctx, subtopicsIDChan, subtopicItemsNilClient, testCensusTopicID)

			Convey("Then no subtopic ids should be sent to subtopicsIDChan channel", func() {
				So(subTopicsIDQuery, ShouldBeEmpty)
			})
		})
	})
}
