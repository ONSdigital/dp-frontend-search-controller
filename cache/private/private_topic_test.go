package private

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-frontend-search-controller/cache"
	"github.com/ONSdigital/dp-topic-api/models"
	"github.com/ONSdigital/dp-topic-api/sdk"
	topicCliErr "github.com/ONSdigital/dp-topic-api/sdk/errors"
	mockTopic "github.com/ONSdigital/dp-topic-api/sdk/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testCensusTitle         = "Census"
	testRootTitle           = "Root Topic"
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
		ID:      cache.CensusTopicID,
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
		ID:          cache.CensusTopicID,
		Title:       testCensusTitle,
		SubtopicIds: &[]string{"5678", "1235"},
	}

	testEconomyRootTopic = models.Topic{
		ID:          "1458",
		Title:       "Economy",
		SubtopicIds: &[]string{},
	}

	testCensusSubTopic1 = models.Topic{
		ID:          testCensusSubTopicID1,
		Title:       "Census Sub 1",
		SubtopicIds: &[]string{"8901"},
	}

	testCensusSubTopic2 = models.Topic{
		ID:          testCensusSubTopicID2,
		Title:       "Census Sub 2",
		SubtopicIds: &[]string{},
	}

	testCensusSubTopic1Sub = models.Topic{
		ID:          testCensusSubSubTopicID,
		Title:       "Census Sub 1 - Sub",
		SubtopicIds: &[]string{},
	}

	expectedCensusTopicCache = &cache.Topic{
		ID:              cache.CensusTopicID,
		LocaliseKeyName: testCensusTitle,
		Query:           fmt.Sprintf("%s,%s,%s", testCensusSubTopicID1, testCensusSubTopicID2, testCensusSubSubTopicID),
	}
)

func mockGetSubtopicsIDsPrivate(ctx context.Context, subtopicsChan chan models.TopicResponse, topicClient sdk.Clienter, topLevelTopicID string) string {
	var rootTopic models.Topic

	switch topLevelTopicID {
	case testCensusSubTopicID2:
		rootTopic = testCensusSubTopic2
	default:
		rootTopic = testCensusRootTopic
	}

	testTopicCache := getRootTopicCachePrivate(ctx, "", subtopicsChan, topicClient, rootTopic)

	return testTopicCache.Query
}

func TestUpdateCensusTopic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockedTopicClient := &mockTopic.ClienterMock{
		GetRootTopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PrivateSubtopics, topicCliErr.Error) {
			return testRootTopicsPrivate, nil
		},

		GetSubtopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PrivateSubtopics, topicCliErr.Error) {
			switch id {
			case cache.CensusTopicID:
				return testCensusSubTopicsPrivate, nil
			case testCensusSubTopicID1:
				return testCensusSubTopic1SubTopicsPrivate, nil
			default:
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			}
		},
	}

	Convey("Given census root topic does exist and has subtopics", t, func() {
		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache := UpdateCensusTopic(ctx, "", mockedTopicClient)()

			Convey("Then the census topic cache is returned", func() {
				So(respCensusTopicCache, ShouldNotBeNil)

				So(respCensusTopicCache.ID, ShouldEqual, expectedCensusTopicCache.ID)
				So(respCensusTopicCache.LocaliseKeyName, ShouldEqual, expectedCensusTopicCache.LocaliseKeyName)

				So(respCensusTopicCache.Query, ShouldContainSubstring, testCensusSubTopicID1)
				So(respCensusTopicCache.Query, ShouldContainSubstring, testCensusSubTopicID2)
				So(respCensusTopicCache.Query, ShouldContainSubstring, testCensusSubSubTopicID)
			})
		})
	})

	Convey("Given an error in getting root topics from topic-api", t, func() {
		failedRootTopicClient := &mockTopic.ClienterMock{
			GetRootTopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PrivateSubtopics, topicCliErr.Error) {
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			},
		}

		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache := UpdateCensusTopic(ctx, "", failedRootTopicClient)()

			Convey("Then an empty census topic cache should be returned", func() {
				So(respCensusTopicCache, ShouldResemble, cache.GetEmptyCensusTopic())
			})
		})
	})

	Convey("Given root topics private items is nil", t, func() {
		rootTopicsNilClient := &mockTopic.ClienterMock{
			GetRootTopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PrivateSubtopics, topicCliErr.Error) {
				rootTopicPrivateItemsNil := *testRootTopicsPrivate
				rootTopicPrivateItemsNil.PrivateItems = nil
				return &rootTopicPrivateItemsNil, nil
			},
		}

		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache := UpdateCensusTopic(ctx, "", rootTopicsNilClient)()

			Convey("Then an empty census topic cache should be returned", func() {
				So(respCensusTopicCache, ShouldResemble, cache.GetEmptyCensusTopic())
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
			GetRootTopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PrivateSubtopics, topicCliErr.Error) {
				return NonCensusRootTopics, nil
			},
		}

		Convey("When UpdateCensusTopicPrivate is called", func() {
			respCensusTopicCache := UpdateCensusTopic(ctx, "", censusTopicNotExistClient)()

			Convey("Then an empty census topic cache should be returned", func() {
				So(respCensusTopicCache, ShouldResemble, cache.GetEmptyCensusTopic())
			})
		})
	})
}

func TestUpdateDataTopics(t *testing.T) {
	ctx := context.Background()
	serviceAuthToken := "test-token"

	expectedTopics := []*cache.Topic{
		{ID: "6734", LocaliseKeyName: "Economy", Query: "6734,1834"},
		{ID: "1834", LocaliseKeyName: "Environmental Accounts"},
		{ID: "1234", LocaliseKeyName: "Business"},
	}
	emptyTopic := []*cache.Topic{cache.GetEmptyTopic()}

	Convey("Given root topics exist and have subtopics", t, func() {
		mockClient := &mockTopic.ClienterMock{
			GetTopicPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, topicID string) (*models.TopicResponse, topicCliErr.Error) {
				switch topicID {
				case cache.RootTopicID:
					return &models.TopicResponse{
						Current: &models.Topic{
							ID:          "9999",
							Title:       "Root",
							SubtopicIds: &[]string{"6734", "1234"},
						},
					}, nil
				case "6734":
					return &models.TopicResponse{
						Current: &models.Topic{
							ID:          "6734",
							Title:       "Economy",
							SubtopicIds: &[]string{"1834"},
						},
					}, nil
				case "1834":
					return &models.TopicResponse{
						Current: &models.Topic{
							ID:          "1834",
							Title:       "Environmental Accounts",
							SubtopicIds: &[]string{},
						},
					}, nil
				case "1234":
					return &models.TopicResponse{
						Current: &models.Topic{
							ID:          "1234",
							Title:       "Business",
							SubtopicIds: &[]string{},
						},
					}, nil
				default:
					return nil, topicCliErr.StatusError{
						Err: errors.New("unexpected error"),
					}
				}
			},
		}

		Convey("When UpdateDataTopicsPrivate is called", func() {
			respTopics := UpdateDataTopics(ctx, serviceAuthToken, mockClient)()

			Convey("Then the topics cache is returned", func() {
				So(respTopics, ShouldNotBeNil)
				So(len(respTopics), ShouldEqual, len(expectedTopics))
				for i, expected := range expectedTopics {
					So(respTopics[i].ID, ShouldEqual, expected.ID)
					So(respTopics[i].LocaliseKeyName, ShouldEqual, expected.LocaliseKeyName)
				}
			})
		})
	})

	Convey("Given an error in getting root topics from topic-api", t, func() {
		mockClient := &mockTopic.ClienterMock{
			GetTopicPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, topicID string) (*models.TopicResponse, topicCliErr.Error) {
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			},
		}

		Convey("When UpdateDataTopicsPrivate is called", func() {
			respTopics := UpdateDataTopics(ctx, serviceAuthToken, mockClient)()

			Convey("Then an empty topic cache should be returned", func() {
				So(respTopics, ShouldResemble, emptyTopic)
			})
		})
	})

	Convey("Given root topics private items is nil", t, func() {
		mockClient := &mockTopic.ClienterMock{
			GetTopicPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, topicID string) (*models.TopicResponse, topicCliErr.Error) {
				return &models.TopicResponse{
					Current: &models.Topic{
						ID:          topicID,
						SubtopicIds: nil,
					},
				}, nil
			},
		}

		Convey("When UpdateDataTopicsPrivate is called", func() {
			respTopics := UpdateDataTopics(ctx, serviceAuthToken, mockClient)()

			Convey("Then an empty topic cache should be returned", func() {
				So(respTopics, ShouldResemble, emptyTopic)
			})
		})
	})

	Convey("Given root topics exist but no data topics found", t, func() {
		mockClient := &mockTopic.ClienterMock{
			GetTopicPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, topicID string) (*models.TopicResponse, topicCliErr.Error) {
				if topicID == cache.RootTopicID {
					return &models.TopicResponse{
						Current: &models.Topic{
							ID:          "9999",
							SubtopicIds: &[]string{},
						},
					}, nil
				}
				return nil, nil
			},
		}

		Convey("When UpdateDataTopicsPrivate is called", func() {
			respTopics := UpdateDataTopics(ctx, serviceAuthToken, mockClient)()

			Convey("Then an empty topic cache should be returned", func() {
				So(respTopics, ShouldResemble, emptyTopic)
			})
		})
	})
}

func TestGetRootTopicCachePrivate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	subtopicsChan := make(chan models.TopicResponse)

	mockedTopicClient := &mockTopic.ClienterMock{
		GetSubtopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PrivateSubtopics, topicCliErr.Error) {
			switch id {
			case cache.CensusTopicID:
				return testCensusSubTopicsPrivate, nil
			case testCensusSubTopicID1:
				return testCensusSubTopic1SubTopicsPrivate, nil
			default:
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			}
		},
	}

	Convey("Given topic has subtopics", t, func() {
		Convey("When getRootTopicCachePrivate is called", func() {
			respCensusTopicCache := getRootTopicCachePrivate(ctx, "", subtopicsChan, mockedTopicClient, testCensusRootTopic)

			Convey("Then the census topic cache is returned", func() {
				So(respCensusTopicCache, ShouldNotBeNil)
				So(respCensusTopicCache.ID, ShouldEqual, expectedCensusTopicCache.ID)
				So(respCensusTopicCache.LocaliseKeyName, ShouldEqual, expectedCensusTopicCache.LocaliseKeyName)

				So(respCensusTopicCache.Query, ShouldContainSubstring, testCensusSubTopicID1)
				So(respCensusTopicCache.Query, ShouldContainSubstring, testCensusSubTopicID2)
				So(respCensusTopicCache.Query, ShouldContainSubstring, testCensusSubSubTopicID)
			})
		})
	})
}

func TestGetSubtopicsIDsPrivate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockedTopicClient := &mockTopic.ClienterMock{
		GetSubtopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PrivateSubtopics, topicCliErr.Error) {
			switch id {
			case cache.CensusTopicID:
				return testCensusSubTopicsPrivate, nil
			case testCensusSubTopicID1:
				return testCensusSubTopic1SubTopicsPrivate, nil
			case testCensusSubTopicID2:
				return nil, topicCliErr.StatusError{
					Err:  errors.New("topic not found"),
					Code: http.StatusNotFound,
				}
			default:
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			}
		},
	}

	Convey("Given topic has subtopics", t, func() {
		subtopicsChan := make(chan models.TopicResponse)

		Convey("When getSubtopicsIDsPrivate is called", func() {
			subTopicsIDQuery := mockGetSubtopicsIDsPrivate(ctx, subtopicsChan, mockedTopicClient, cache.CensusTopicID)

			Convey("Then subtopic ids should be sent to subtopicsIDChan channel", func() {
				So(subTopicsIDQuery, ShouldNotBeEmpty)
				So(subTopicsIDQuery, ShouldContainSubstring, testCensusSubTopicID1)
				So(subTopicsIDQuery, ShouldContainSubstring, testCensusSubTopicID2)
				So(subTopicsIDQuery, ShouldContainSubstring, testCensusSubSubTopicID)
			})
		})
	})

	Convey("Given topic has no subtopics", t, func() {
		subtopicsChan := make(chan models.TopicResponse)

		Convey("When getSubtopicsIDsPrivate is called", func() {
			subTopicsIDQuery := mockGetSubtopicsIDsPrivate(ctx, subtopicsChan, mockedTopicClient, testCensusSubTopicID2)

			Convey("Then no subtopic ids should be sent to subtopicsIDChan channel", func() {
				// the query only contains the root topic id and no subtopic ids
				So(subTopicsIDQuery, ShouldEqual, testCensusSubTopicID2)
			})
		})
	})

	Convey("Given an error in getting sub topics from topic-api", t, func() {
		subtopicsChan := make(chan models.TopicResponse)

		failedGetSubtopicClient := &mockTopic.ClienterMock{
			GetSubtopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PrivateSubtopics, topicCliErr.Error) {
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			},
		}

		Convey("When getSubtopicsIDsPrivate is called", func() {
			subTopicsIDQuery := mockGetSubtopicsIDsPrivate(ctx, subtopicsChan, failedGetSubtopicClient, cache.CensusTopicID)

			Convey("Then no subtopic ids should be sent to subtopicsIDChan channel", func() {
				// the query only contains the root topic id and no subtopic ids
				So(subTopicsIDQuery, ShouldEqual, cache.CensusTopicID)
			})
		})
	})

	Convey("Given sub topics private items is nil", t, func() {
		subtopicsChan := make(chan models.TopicResponse)

		subtopicItemsNilClient := &mockTopic.ClienterMock{
			GetSubtopicsPrivateFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PrivateSubtopics, topicCliErr.Error) {
				topicItemsNil := *testCensusSubTopicsPrivate
				topicItemsNil.PrivateItems = nil
				return &topicItemsNil, nil
			},
		}

		Convey("When getSubtopicsIDsPrivate is called", func() {
			subTopicsIDQuery := mockGetSubtopicsIDsPrivate(ctx, subtopicsChan, subtopicItemsNilClient, cache.CensusTopicID)

			Convey("Then no subtopic ids should be sent to subtopicsIDChan channel", func() {
				// the query only contains the root topic id and no subtopic ids
				So(subTopicsIDQuery, ShouldEqual, cache.CensusTopicID)
			})
		})
	})
}
