package mapper

import (
	"embed"
	"encoding/json"

	searchModels "github.com/ONSdigital/dp-search-api/models"

	zebedeeC "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
)

//go:embed data/*
var res embed.FS

// GetMockSearchResponse get a mock search response in searchC.Response type from dp-search-api using ES 7.10
func GetMockSearchResponse() (*searchModels.SearchResponse, error) {
	var respC *searchModels.SearchResponse

	sampleResponse, err := res.ReadFile("data/mock_search_response.json")
	if err != nil {
		return &searchModels.SearchResponse{}, err
	}

	err = json.Unmarshal(sampleResponse, &respC)
	if err != nil {
		return &searchModels.SearchResponse{}, err
	}

	return respC, nil
}

// GetMockHomepageContent gets mock homepage content
func GetMockHomepageContent() (zebedeeC.HomepageContent, error) {
	var hc zebedeeC.HomepageContent

	mockContent, err := res.ReadFile("data/mock_homepage_content.json")
	if err != nil {
		return zebedeeC.HomepageContent{}, err
	}

	err = json.Unmarshal(mockContent, &hc)
	if err != nil {
		return zebedeeC.HomepageContent{}, err
	}

	return hc, nil
}

func GetFindADatasetResponse() (*searchModels.SearchResponse, error) {
	var respC *searchModels.SearchResponse

	sampleResponse, err := res.ReadFile("data/mock_find_a_dataset_response.json")
	if err != nil {
		return &searchModels.SearchResponse{}, err
	}

	err = json.Unmarshal(sampleResponse, &respC)
	if err != nil {
		return &searchModels.SearchResponse{}, err
	}

	return respC, nil
}

func GetMockZebedeePageDataResponse() (zebedeeC.PageData, error) {
	var hc zebedeeC.PageData

	mockContent, err := res.ReadFile("data/mock_zebedee_page_content.json")
	if err != nil {
		return zebedeeC.PageData{}, err
	}

	err = json.Unmarshal(mockContent, &hc)
	if err != nil {
		return zebedeeC.PageData{}, err
	}

	return hc, nil
}
