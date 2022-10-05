package mapper

import (
	"encoding/json"
	"os"

	searchC "github.com/ONSdigital/dp-api-clients-go/v2/site-search"
	zebedeeC "github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
)

// GetMockSearchResponse get a mock search response in searchC.Response type from dp-search-api using ES 7.10
func GetMockSearchResponse() (searchC.Response, error) {
	var respC searchC.Response

	sampleResponse, err := os.ReadFile("../mapper/data/mock_search_response.json")
	if err != nil {
		return searchC.Response{}, err
	}

	err = json.Unmarshal(sampleResponse, &respC)
	if err != nil {
		return searchC.Response{}, err
	}

	return respC, nil
}

// GetMockHomepageContent gets mock homepage content
func GetMockHomepageContent() (zebedeeC.HomepageContent, error) {
	var hc zebedeeC.HomepageContent

	mockContent, err := os.ReadFile("../mapper/data/mock_homepage_content.json")
	if err != nil {
		return zebedeeC.HomepageContent{}, err
	}

	err = json.Unmarshal(mockContent, &hc)
	if err != nil {
		return zebedeeC.HomepageContent{}, err
	}

	return hc, nil
}
