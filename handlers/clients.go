package handlers

import (
	"context"
	"io"
	"net/url"

	searchCli "github.com/ONSdigital/dp-api-clients-go/v2/site-search"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	coreModel "github.com/ONSdigital/dp-renderer/model"
)

//go:generate moq -out clients_mock.go -pkg handlers . RenderClient SearchClient ZebedeeClient

// ClientError is an interface that can be used to retrieve the status code if a client has errored
type ClientError interface {
	Code() int
}

// RenderClient is an interface with methods for require for rendering a template
type RenderClient interface {
	BuildPage(w io.Writer, pageModel interface{}, templateName string)
	NewBasePageModel() coreModel.Page
}

// SearchClient is an interface with methods required for a search client
type SearchClient interface {
	GetSearch(ctx context.Context, userAuthToken, serviceAuthToken, collectionID string, query url.Values) (r searchCli.Response, err error)
}

// ZebedeeClient is an interface with methods required for a zebedee client
type ZebedeeClient interface {
	GetHomepageContent(ctx context.Context, userAuthToken, collectionID, lang, path string) (m zebedee.HomepageContent, err error)
}
