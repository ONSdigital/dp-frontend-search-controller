package mapper

import (
	"encoding/json"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"io/ioutil"

	searchC "github.com/ONSdigital/dp-api-clients-go/v2/site-search"
)

// GetMockSearchResponse get a mock search response in searchC.Response type
func GetMockSearchResponse() (searchC.Response, error) {
	var respC searchC.Response

	sampleResponse, err := ioutil.ReadFile("../mapper/data/mock_search_response.json")
	if err != nil {
		return searchC.Response{}, err
	}

	err = json.Unmarshal(sampleResponse, &respC)
	if err != nil {
		return searchC.Response{}, err
	}

	return respC, nil
}

// GetMockDepartmentResponse get a mock department response in searchC.Department type
func GetMockDepartmentResponse() (searchC.Department, error) {
	var respC searchC.Department

	sampleResponse, err := ioutil.ReadFile("../mapper/data/mock_department_response.json")
	if err != nil {
		return searchC.Department{}, err
	}

	err = json.Unmarshal(sampleResponse, &respC)
	if err != nil {
		return searchC.Department{}, err
	}

	return respC, nil
}

func GetMockHomepageContent() zebedee.HomepageContent {
	return zebedee.HomepageContent{

	}
}


//func GetTestEmergencyBanner() zebedee.EmergencyBanner {
//	return zebedee.EmergencyBanner{
//		Type:        "notable_death",
//		Title:       "This is not not an emergency",
//		Description: "Something has gone wrong",
//		URI:         "google.com",
//		LinkText:    "More info",
//	}
//}
//
//func GetTestServiceMessage() string {
//	return "Test service message"
//}
