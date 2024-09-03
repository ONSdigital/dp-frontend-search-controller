package handlers

import (
	"context"
	"io"

	searchModels "github.com/ONSdigital/dp-search-api/models"
	searchSDK "github.com/ONSdigital/dp-search-api/sdk"
	searchError "github.com/ONSdigital/dp-search-api/sdk/errors"

	topicModels "github.com/ONSdigital/dp-topic-api/models"
	topicSDK "github.com/ONSdigital/dp-topic-api/sdk"
	topicError "github.com/ONSdigital/dp-topic-api/sdk/errors"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	coreModel "github.com/ONSdigital/dp-renderer/v2/model"
)

//go:generate moq -out clients_mock.go -pkg handlers . RenderClient SearchClient ZebedeeClient TopicClient

// ClientError is an interface that can be used to retrieve the status code if a client has errored
type ClientError interface {
	Code() int
}

// Search API returns a SearchClientError interface, which differs from ZebedeeClient
type SearchClientError interface {
	Status() int
}

// RenderClient is an interface with methods for require for rendering a template
type RenderClient interface {
	BuildPage(w io.Writer, pageModel interface{}, templateName string)
	NewBasePageModel() coreModel.Page
}

// SearchClient is an interface with methods required for a search client
type SearchClient interface {
	GetSearch(ctx context.Context, options searchSDK.Options) (*searchModels.SearchResponse, searchError.Error)
}

// ZebedeeClient is an interface with methods required for a zebedee client
type ZebedeeClient interface {
	GetHomepageContent(ctx context.Context, userAuthToken, collectionID, lang, path string) (m zebedee.HomepageContent, err error)
	GetPageData(ctx context.Context, userAuthToken, collectionID, lang, path string) (m zebedee.PageData, err error)
}

// TopicClient is an interface with methods required for a zebedee client
type TopicClient interface {
	GetNavigationPublic(ctx context.Context, reqHeaders topicSDK.Headers, options topicSDK.Options) (*topicModels.Navigation, topicError.Error)
	GetRootTopicsPrivate(ctx context.Context, reqHeaders topicSDK.Headers) (*topicModels.PrivateSubtopics, topicError.Error)
	GetRootTopicsPublic(ctx context.Context, reqHeaders topicSDK.Headers) (*topicModels.PublicSubtopics, topicError.Error)
	GetSubtopicsPrivate(ctx context.Context, reqHeaders topicSDK.Headers, id string) (*topicModels.PrivateSubtopics, topicError.Error)
	GetSubtopicsPublic(ctx context.Context, reqHeaders topicSDK.Headers, id string) (*topicModels.PublicSubtopics, topicError.Error)
	GetTopicPrivate(ctx context.Context, reqHeaders topicSDK.Headers, id string) (*topicModels.TopicResponse, topicError.Error)
	GetTopicPublic(ctx context.Context, reqHeaders topicSDK.Headers, id string) (*topicModels.Topic, topicError.Error)
}
