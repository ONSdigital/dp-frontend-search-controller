package public

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
	mockTopicCli "github.com/ONSdigital/dp-topic-api/sdk/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

const (
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
		ID:          cache.CensusTopicID,
		Title:       testCensusTitle,
		SubtopicIds: &[]string{"5678", "1235"},
	}

	testEconomyRootTopic = models.Topic{
		ID:          "6734",
		Title:       "Economy",
		SubtopicIds: &[]string{"1834"},
	}

	testEconomyRootTopic2 = models.Topic{
		ID:          "6734",
		Title:       "Economy",
		SubtopicIds: &[]string{"1834", "1234"},
	}

	testBusinessRootTopic = models.Topic{
		ID:          "1234",
		Title:       "Business",
		SubtopicIds: &[]string{},
	}

	testBusinessRootTopic2 = models.Topic{
		ID:          "1234",
		Title:       "Business",
		SubtopicIds: &[]string{"6734"},
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
		SubtopicIds: &[]string{"8901"},
	}

	testCensusSubTopic2 = models.Topic{
		ID:          testCensusSubTopicID2,
		Title:       "Census Sub 2",
		SubtopicIds: &[]string{},
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
		SubtopicIds: &[]string{},
	}

	expectedCensusTopicCache = &cache.Topic{
		ID:              cache.CensusTopicID,
		LocaliseKeyName: testCensusTitle,
		Query:           fmt.Sprintf("%s,%s,%s", testCensusSubTopicID1, testCensusSubTopicID2, testCensusSubSubTopicID),
	}
)

func mockGetSubtopicsIDsPublic(ctx context.Context, subtopicsChan chan models.Topic, topicClient sdk.Clienter, topLevelTopicID string) string {
	var rootTopic models.Topic

	switch topLevelTopicID {
	case testCensusSubTopicID2:
		rootTopic = testCensusSubTopic2
	default:
		rootTopic = testCensusRootTopic
	}

	testTopicCache := getRootTopicCachePublic(ctx, topicClient, rootTopic)

	return testTopicCache.Query
}

func TestUpdateCensusTopic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockedTopicClient := &mockTopicCli.ClienterMock{
		GetRootTopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PublicSubtopics, topicCliErr.Error) {
			return testRootTopics, nil
		},

		GetSubtopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PublicSubtopics, topicCliErr.Error) {
			switch id {
			case cache.CensusTopicID:
				return testCensusSubTopics, nil
			case testCensusSubTopicID1:
				return testCensusSubTopic1SubTopics, nil
			default:
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			}
		},
	}

	Convey("Given census root topic does exist and has subtopics", t, func() {
		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache := UpdateCensusTopic(ctx, mockedTopicClient)()

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
		failedRootTopicClient := &mockTopicCli.ClienterMock{
			GetRootTopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PublicSubtopics, topicCliErr.Error) {
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			},
		}

		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache := UpdateCensusTopic(ctx, failedRootTopicClient)()

			Convey("Then an empty census topic cache should be returned", func() {
				So(respCensusTopicCache, ShouldResemble, cache.GetEmptyCensusTopic())
			})
		})
	})

	Convey("Given root topics public items is nil", t, func() {
		rootTopicsNilClient := &mockTopicCli.ClienterMock{
			GetRootTopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PublicSubtopics, topicCliErr.Error) {
				rootTopicPublicItemsNil := *testRootTopics
				rootTopicPublicItemsNil.PublicItems = nil
				return &rootTopicPublicItemsNil, nil
			},
		}

		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache := UpdateCensusTopic(ctx, rootTopicsNilClient)()

			Convey("Then an empty census topic cache should be returned", func() {
				So(respCensusTopicCache, ShouldResemble, cache.GetEmptyCensusTopic())
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
			GetRootTopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PublicSubtopics, topicCliErr.Error) {
				return NonCensusRootTopics, nil
			},
		}

		Convey("When UpdateCensusTopic is called", func() {
			respCensusTopicCache := UpdateCensusTopic(ctx, censusTopicNotExistClient)()

			Convey("Then an empty census topic cache should be returned", func() {
				So(respCensusTopicCache, ShouldResemble, cache.GetEmptyCensusTopic())
			})
		})
	})
}

func TestUpdateDataTopics(t *testing.T) {
	ctx := context.Background()

	expectedTopics := []*cache.Topic{
		{ID: "6734", LocaliseKeyName: "Economy", Query: "6734,1834"},
		{ID: "1834", LocaliseKeyName: "Environmental Accounts"},
		{ID: "1234", LocaliseKeyName: "Business"},
	}
	emptyTopic := []*cache.Topic{cache.GetEmptyTopic()}

	Convey("Given root topics exist and have subtopics with no duplicate topics", t, func() {
		mockClient := &mockTopicCli.ClienterMock{
			GetRootTopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PublicSubtopics, topicCliErr.Error) {
				return &models.PublicSubtopics{
					Count:       2,
					Offset:      0,
					Limit:       50,
					TotalCount:  2,
					PublicItems: &[]models.Topic{testEconomyRootTopic, testBusinessRootTopic},
				}, nil
			},
			GetTopicPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, topicID string) (*models.Topic, topicCliErr.Error) {
				switch topicID {
				case "9999":
					return &models.Topic{
						ID: "9999", Title: "Root", SubtopicIds: &[]string{"6734", "1234", "1834"},
					}, nil
				case "6734":
					return &models.Topic{
						ID: "6734", Title: "Economy", SubtopicIds: &[]string{"1834"},
					}, nil
				case "1834":
					return &models.Topic{
						ID: "1834", Title: "Environmental Accounts", SubtopicIds: &[]string{},
					}, nil
				case "1234":
					return &models.Topic{
						ID: "1234", Title: "Business", SubtopicIds: &[]string{},
					}, nil
				default:
					return nil, topicCliErr.StatusError{
						Err: errors.New("unexpected error"),
					}
				}
			},
		}

		Convey("When UpdateDataTopics is called", func() {
			respTopics := UpdateDataTopics(ctx, mockClient)()

			Convey("Then the topics cache is returned as expected", func() {
				So(respTopics, ShouldNotBeNil)
				So(len(respTopics), ShouldEqual, len(expectedTopics))
				for i, expected := range expectedTopics {
					So(respTopics[i].ID, ShouldEqual, expected.ID)
					So(respTopics[i].LocaliseKeyName, ShouldEqual, expected.LocaliseKeyName)
				}
			})
		})
	})

	Convey("Given root topics exist and have subtopics with duplicate topics", t, func() {
		mockClient := &mockTopicCli.ClienterMock{
			GetRootTopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PublicSubtopics, topicCliErr.Error) {
				return &models.PublicSubtopics{
					Count:       2,
					Offset:      0,
					Limit:       50,
					TotalCount:  2,
					PublicItems: &[]models.Topic{testEconomyRootTopic, testBusinessRootTopic, testEconomyRootTopic, testBusinessRootTopic},
				}, nil
			},
			GetTopicPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, topicID string) (*models.Topic, topicCliErr.Error) {
				switch topicID {
				case "9999":
					return &models.Topic{
						ID: "9999", Title: "Root", SubtopicIds: &[]string{"6734", "1234"},
					}, nil
				case "6734":
					return &models.Topic{
						ID: "6734", Title: "Economy", SubtopicIds: &[]string{"1834"},
					}, nil
				case "1834":
					return &models.Topic{
						ID: "1834", Title: "Environmental Accounts", SubtopicIds: &[]string{},
					}, nil
				case "1234":
					return &models.Topic{
						ID: "1234", Title: "Business", SubtopicIds: &[]string{},
					}, nil
				default:
					return nil, topicCliErr.StatusError{
						Err: errors.New("unexpected error"),
					}
				}
			},
		}

		Convey("When UpdateDataTopics is called", func() {
			respTopics := UpdateDataTopics(ctx, mockClient)()

			Convey("Then the topics cache is returned with expected number of topics excluding duplicates", func() {
				So(respTopics, ShouldNotBeNil)
				So(len(respTopics), ShouldEqual, len(expectedTopics))
				for i, expected := range expectedTopics {
					So(respTopics[i].ID, ShouldEqual, expected.ID)
					So(respTopics[i].LocaliseKeyName, ShouldEqual, expected.LocaliseKeyName)
				}
			})
		})
	})

	Convey("Given root topics exist and have subtopics with a loop of topics", t, func() {
		mockClient := &mockTopicCli.ClienterMock{
			GetRootTopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PublicSubtopics, topicCliErr.Error) {
				return &models.PublicSubtopics{
					Count:       2,
					Offset:      0,
					Limit:       50,
					TotalCount:  2,
					PublicItems: &[]models.Topic{testEconomyRootTopic, testEconomyRootTopic2, testBusinessRootTopic2},
				}, nil
			},
			GetTopicPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, topicID string) (*models.Topic, topicCliErr.Error) {
				switch topicID {
				case "9999":
					return &models.Topic{
						ID: "9999", Title: "Root", SubtopicIds: &[]string{"6734", "1234"},
					}, nil
				case "6734":
					return &models.Topic{
						ID: "6734", Title: "Economy", SubtopicIds: &[]string{"1834", "1234"},
					}, nil
				case "1834":
					return &models.Topic{
						ID: "1834", Title: "Environmental Accounts", SubtopicIds: &[]string{},
					}, nil
				case "1234":
					return &models.Topic{
						ID: "1234", Title: "Business", SubtopicIds: &[]string{"6734"},
					}, nil
				default:
					return nil, topicCliErr.StatusError{
						Err: errors.New("unexpected error"),
					}
				}
			},
		}

		Convey("When UpdateDataTopics is called", func() {
			respTopics := UpdateDataTopics(ctx, mockClient)()

			Convey("Then the topics cache is returned as expected with no duplicates and does not get stuck in a loop", func() {
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
		mockClient := &mockTopicCli.ClienterMock{
			GetRootTopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PublicSubtopics, topicCliErr.Error) {
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			},
		}

		Convey("When UpdateDataTopics is called", func() {
			respTopics := UpdateDataTopics(ctx, mockClient)()

			Convey("Then an empty topic cache should be returned", func() {
				So(respTopics, ShouldResemble, emptyTopic)
			})
		})
	})

	Convey("Given root topics public items is nil", t, func() {
		mockClient := &mockTopicCli.ClienterMock{
			GetRootTopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PublicSubtopics, topicCliErr.Error) {
				return &models.PublicSubtopics{
					Count:       0,
					Offset:      0,
					Limit:       50,
					TotalCount:  0,
					PublicItems: nil,
				}, nil
			},
			GetTopicPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, topicID string) (*models.Topic, topicCliErr.Error) {
				return &models.Topic{ID: topicID, SubtopicIds: nil}, nil
			},
		}

		Convey("When UpdateDataTopics is called", func() {
			respTopics := UpdateDataTopics(ctx, mockClient)()

			Convey("Then an empty topic cache should be returned", func() {
				So(respTopics, ShouldResemble, emptyTopic)
			})
		})
	})

	Convey("Given root topics exist but no data topics found", t, func() {
		mockClient := &mockTopicCli.ClienterMock{
			GetRootTopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers) (*models.PublicSubtopics, topicCliErr.Error) {
				return &models.PublicSubtopics{
					Count:       0,
					Offset:      0,
					Limit:       50,
					TotalCount:  0,
					PublicItems: &[]models.Topic{},
				}, nil
			},
		}

		Convey("When UpdateDataTopics is called", func() {
			respTopics := UpdateDataTopics(ctx, mockClient)()

			Convey("Then an empty topic cache should be returned", func() {
				So(respTopics, ShouldResemble, emptyTopic)
			})
		})
	})
}

func TestGetRootTopicCachePublic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockedTopicClient := &mockTopicCli.ClienterMock{
		GetSubtopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PublicSubtopics, topicCliErr.Error) {
			switch id {
			case cache.CensusTopicID:
				return testCensusSubTopics, nil
			case testCensusSubTopicID1:
				return testCensusSubTopic1SubTopics, nil
			default:
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			}
		},
	}

	Convey("Given topic has subtopics", t, func() {
		Convey("When getRootTopicCache is called", func() {
			respCensusTopicCache := getRootTopicCachePublic(ctx, mockedTopicClient, testCensusRootTopic)

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

func TestGetSubtopicsIDsPublic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	mockedTopicClient := &mockTopicCli.ClienterMock{
		GetSubtopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PublicSubtopics, topicCliErr.Error) {
			switch id {
			case cache.CensusTopicID:
				return testCensusSubTopics, nil
			case testCensusSubTopicID1:
				return testCensusSubTopic1SubTopics, nil
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
		subtopicsChan := make(chan models.Topic)

		Convey("When getSubtopicsIDsPublic is called", func() {
			subTopicsIDQuery := mockGetSubtopicsIDsPublic(ctx, subtopicsChan, mockedTopicClient, cache.CensusTopicID)

			Convey("Then subtopic ids should be sent to subtopicsIDChan channel", func() {
				So(subTopicsIDQuery, ShouldNotBeEmpty)
				So(subTopicsIDQuery, ShouldContainSubstring, testCensusSubTopicID1)
				So(subTopicsIDQuery, ShouldContainSubstring, testCensusSubTopicID2)
				So(subTopicsIDQuery, ShouldContainSubstring, testCensusSubSubTopicID)
			})
		})
	})

	Convey("Given topic has no subtopics", t, func() {
		subtopicsChan := make(chan models.Topic)

		Convey("When getSubtopicsIDsPublic is called", func() {
			subTopicsIDQuery := mockGetSubtopicsIDsPublic(ctx, subtopicsChan, mockedTopicClient, testCensusSubTopicID2)

			Convey("Then no subtopic ids should be sent to subtopicsIDChan channel", func() {
				// the query only contains the root topic id and no subtopic ids
				So(subTopicsIDQuery, ShouldEqual, testCensusSubTopicID2)
			})
		})
	})

	Convey("Given an error in getting sub topics from topic-api", t, func() {
		subtopicsChan := make(chan models.Topic)

		failedGetSubtopicClient := &mockTopicCli.ClienterMock{
			GetSubtopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PublicSubtopics, topicCliErr.Error) {
				return nil, topicCliErr.StatusError{
					Err: errors.New("unexpected error"),
				}
			},
		}

		Convey("When getSubtopicsIDsPublic is called", func() {
			subTopicsIDQuery := mockGetSubtopicsIDsPublic(ctx, subtopicsChan, failedGetSubtopicClient, cache.CensusTopicID)

			Convey("Then no subtopic ids should be sent to subtopicsIDChan channel", func() {
				// the query only contains the root topic id and no subtopic ids
				So(subTopicsIDQuery, ShouldEqual, cache.CensusTopicID)
			})
		})
	})

	Convey("Given sub topics public items is nil", t, func() {
		subtopicsChan := make(chan models.Topic)

		subtopicItemsNilClient := &mockTopicCli.ClienterMock{
			GetSubtopicsPublicFunc: func(ctx context.Context, reqHeaders sdk.Headers, id string) (*models.PublicSubtopics, topicCliErr.Error) {
				topicItemsNil := *testCensusSubTopics
				topicItemsNil.PublicItems = nil
				return &topicItemsNil, nil
			},
		}

		Convey("When getSubtopicsIDsPublic is called", func() {
			subTopicsIDQuery := mockGetSubtopicsIDsPublic(ctx, subtopicsChan, subtopicItemsNilClient, cache.CensusTopicID)

			Convey("Then no subtopic ids should be sent to subtopicsIDChan channel", func() {
				// the query only contains the root topic id and no subtopic ids
				So(subTopicsIDQuery, ShouldEqual, cache.CensusTopicID)
			})
		})
	})
}
