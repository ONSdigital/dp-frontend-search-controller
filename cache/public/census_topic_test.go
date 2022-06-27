package public

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	topicCliErr "github.com/ONSdigital/dp-topic-api/apierrors"
	"github.com/ONSdigital/dp-topic-api/models"
	"github.com/ONSdigital/dp-topic-api/sdk"
	mockTopicCli "github.com/ONSdigital/dp-topic-api/sdk/mocks"
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

	expectedCensusTopicCache = &cache.Topic{
		ID:               testCensusTopicID,
		LocaliseKeyName:  testCensusTitle,
		SubtopicsIDQuery: fmt.Sprintf("%s,%s,%s", testCensusSubTopicID1, testCensusSubTopicID2, testCensusSubSubTopicID),
	}
)

func mockGetSubtopicsIDsPublic(ctx context.Context, subtopicsIDChan chan string, topicClient sdk.Clienter, topLevelTopicID string) string {
	var rootTopic models.Topic

	switch topLevelTopicID {
	case testCensusSubTopicID2:
		rootTopic = testCensusSubTopic2
	default:
		rootTopic = testCensusRootTopic
	}

	testTopicCache := getRootTopicCachePublic(ctx, subtopicsIDChan, topicClient, rootTopic)

	return testTopicCache.SubtopicsIDQuery
}

func TestUpdateCensusTopic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockedTopicClient := &mockTopicCli.ClienterMock{
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
		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache, err := UpdateCensusTopic(ctx, mockedTopicClient)()

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
		failedRootTopicClient := &mockTopicCli.ClienterMock{
			GetRootTopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PublicSubtopics, error) {
				return nil, errors.New("unexpected error")
			},
		}

		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache, err := UpdateCensusTopic(ctx, failedRootTopicClient)()

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And the census topic cache returned is nil", func() {
					So(respCensusTopicCache, ShouldBeNil)
				})
			})
		})
	})

	Convey("Given root topics public items is nil", t, func() {
		rootTopicsNilClient := &mockTopicCli.ClienterMock{
			GetRootTopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PublicSubtopics, error) {
				rootTopicPublicItemsNil := *testRootTopics
				rootTopicPublicItemsNil.PublicItems = nil
				return &rootTopicPublicItemsNil, nil
			},
		}

		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache, err := UpdateCensusTopic(ctx, rootTopicsNilClient)()

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

		censusTopicNotExistClient := &mockTopicCli.ClienterMock{
			GetRootTopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PublicSubtopics, error) {
				return NonCensusRootTopics, nil
			},
		}

		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache, err := UpdateCensusTopic(ctx, censusTopicNotExistClient)()

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

	mockedTopicClient := &mockTopicCli.ClienterMock{
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

		Convey("When getRootTopicCache is called", func() {
			respCensusTopicCache := getRootTopicCachePublic(ctx, subtopicsIDChan, mockedTopicClient, testCensusRootTopic)

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

func TestGetSubtopicsIDsPublic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockedTopicClient := &mockTopicCli.ClienterMock{
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
			subTopicsIDQuery := mockGetSubtopicsIDsPublic(ctx, subtopicsIDChan, mockedTopicClient, testCensusTopicID)

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

		Convey("When getSubtopicsIDsPublic is called", func() {
			subTopicsIDQuery := mockGetSubtopicsIDsPublic(ctx, subtopicsIDChan, mockedTopicClient, testCensusSubTopicID2)

			Convey("Then no subtopic ids should be sent to subtopicsIDChan channel", func() {
				So(subTopicsIDQuery, ShouldBeEmpty)
			})
		})
	})

	Convey("Given an error in getting sub topics from topic-api", t, func() {
		subtopicsIDChan := make(chan string)

		failedGetSubtopicClient := &mockTopicCli.ClienterMock{
			GetSubtopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PublicSubtopics, error) {
				return nil, errors.New("unexpected error")
			},
		}

		Convey("When getSubtopicsIDsPublic is called", func() {
			subTopicsIDQuery := mockGetSubtopicsIDsPublic(ctx, subtopicsIDChan, failedGetSubtopicClient, testCensusTopicID)

			Convey("Then no subtopic ids should be sent to subtopicsIDChan channel", func() {
				So(subTopicsIDQuery, ShouldBeEmpty)
			})
		})
	})

	Convey("Given sub topics public items is nil", t, func() {
		subtopicsIDChan := make(chan string)

		subtopicItemsNilClient := &mockTopicCli.ClienterMock{
			GetSubtopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PublicSubtopics, error) {
				topicItemsNil := *testCensusSubTopics
				topicItemsNil.PublicItems = nil
				return &topicItemsNil, nil
			},
		}

		Convey("When getSubtopicsIDsPublic is called", func() {
			subTopicsIDQuery := mockGetSubtopicsIDsPublic(ctx, subtopicsIDChan, subtopicItemsNilClient, testCensusTopicID)

			Convey("Then no subtopic ids should be sent to subtopicsIDChan channel", func() {
				So(subTopicsIDQuery, ShouldBeEmpty)
			})
		})
	})
}
