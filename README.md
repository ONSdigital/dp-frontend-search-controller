# dp-frontend-search-controller

An HTTP service for the controlling of search API

## Getting started

* Run `make debug`

Run new search in the UI:
* run the [web journey](https://github.com/ONSdigital/dp/blob/main/guides/INSTALLING.md#web-journey)

* set `SearchRoutesEnabled` to `true` and `SearchABTestPercentage` to `100` in the config of the [dp-frontend-router](https://github.com/ONSdigital/dp-frontend-router)

## Dependencies

* generate default content with [dp-zebedee-content](https://github.com/ONSdigital/dp-zebedee-content#dp-zebedee-content)
* [dp-design-system](https://github.com/ONSdigital/dp-design-system)
* [dp-search-api](https://github.com/ONSdigital/dp-search-api)
* [dp-topic-api](https://github.com/ONSdigital/dp-topic-api)
* Run Elasticsearch (version 7.10 runs on port 11200) via [dp-compose](https://github.com/ONSdigital/dp-compose) 

No further dependencies other than those defined in `go.mod`

## Configuration

| Environment variable                        | Default                   |                                                                                                                    |
|---------------------------------------------|---------------------------|--------------------------------------------------------------------------------------------------------------------|
| API_ROUTER_URL                              | http://localhost:23200/v1 | The URL of the [dp-api-router](https://github.com/ONSdigital/dp-api-router)                                        |
| BIND_ADDR                                   | :25000                    | The port to bind to                                                                                                |
| CACHE_CENSUS_TOPICS_UPDATE_INTERVAL         | 30m                       | The time interval to update cache for census topics (`time.Duration` format)                                       |
| CACHE_NAVIGATION_UPDATE_INTERVAL            | 30m                       | The time interval to update cache for navigation bar (`time.Duration` format)                                      |
| CENSUS_TOPIC_ID                             | 4445                      | Unique identifer for the census topic, used to get census topics from Topics API                                   |
| DEBUG                                       | false                     | Enable debug mode                                                                                                  |
| DEFAULT_DATASET_SORT                        | release_date              | The default sort for census dataset finder                                                                         |
| DEFAULT_LIMIT                               | 10                        | The default limit of search results in a page                                                                      |
| DEFAULT_MAXIMUM_LIMIT                       | 50                        | The default maximum limit of search results in a page                                                              |
| DEFAULT_MAXIMUM_SEARCH_RESULTS              | 500                       | The default maximum search results                                                                                 |
| DEFAULT_OFFSET                              | 0                         | The default offset of search results                                                                               |
| DEFAULT_PAGE                                | 1                         | The default current page of search results                                                                         |
| DEFAULT_SORT                                | relevance                 | The default sort of search results |
| ENABLE_REWORKED_DATA_AGGREGATION_PAGES      | false                     | Enable the reworked data aggregation pages  |
| ENABLE_CENSUS_DIMENSIONS_FILTER_OPTION      | false                     | Enable dimensions filter for census dataset finder                                                                 |
| ENABLE_CENSUS_POPULATION_TYPE_FILTER_OPTION | false                     | Enable populations filter for census dataset finder                                                                |
| ENABLE_CENSUS_TOPIC_FILTER_OPTION           | false                     | Enable filtering on various census topics                                                                          |
| ENABLE_NEW_NAV_BAR                          | false                     | Enable new dynamic navigation bar                                                                                  |
| GRACEFUL_SHUTDOWN_TIMEOUT                   | 5s                        | The graceful shutdown timeout in seconds (`time.Duration` format)                                                  |
| HEALTHCHECK_CRITICAL_TIMEOUT                | 90s                       | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format) |
| HEALTHCHECK_INTERVAL                        | 30s                       | Time between self-healthchecks (`time.Duration` format)                                                            |
| IS_PUBLISHING                               | false                     | Mode in which service is running                                                                                   |
| PATTERN_LIBRARY_ASSETS_PATH                 | ""                        | Pattern library location                                                                                           |
| SERVICE_AUTH_TOKEN                          | ""                        | This is required to identify the controller when it calls the topic API via the API router in publishing mode      |
| SITE_DOMAIN                                 | localhost                 |                                                                                                                    |
| SUPPORTED_LANGUAGES                         | [2]string{"en", "cy"}     | Supported languages                                                                                                |

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

Copyright © 2020 - 2022, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.