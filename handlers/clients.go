package handlers

import (
	"context"
	"net/url"

	searchCli "github.com/ONSdigital/dp-api-clients-go/site-search"
)

//go:generate moq -out clients_mock.go -pkg handlers . RenderClient SearchClient

// ClientError is an interface that can be used to retrieve the status code if a client has errored
type ClientError interface {
	Code() int
}

// RenderClient is an interface with methods for require for rendering a template
type RenderClient interface {
	Do(string, []byte) ([]byte, error)
}

// SearchClient is an interface with methods required for a search client
type SearchClient interface {
	GetSearch(ctx context.Context, query url.Values) (r searchCli.Response, err error)
	GetDepartments(ctx context.Context, query url.Values) (d searchCli.Department, err error)
}
