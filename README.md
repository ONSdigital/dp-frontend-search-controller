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
* Run Elasticsearch (version 2.4.2 runs on port 9200) via [dp-compose](https://github.com/ONSdigital/dp-compose) 

No further dependencies other than those defined in `go.mod`

## Configuration

| Environment variable              | Default                      | Description
| --------------------------------- | ---------------------------- | --------------------------------------------------
| API_ROUTER_URL                    | http://localhost:23200/v1    | The URL of the [dp-api-router](https://github.com/ONSdigital/dp-api-router)
| BIND_ADDR                         | localhost:25000              | The host and port to bind to
| DEBUG                             | false                        | Enable debug mode
| DEFAULT_LIMIT                     | 10                           | The default limit of search results in a page
| DEFAULT_MAXIMUM_LIMIT             | 50                           | The default maximum limit of search results in a page
| DEFAULT_MAXIMUM_SEARCH_RESULTS    | 500                          | The default maximum search results
| DEFAULT_OFFSET                    | 0                            | The default offset of search results
| DEFAULT_PAGE                      | 1                            | The default current page of search results
| DEFAULT_SORT                      | relevance                    | The default sort of search results
| ENABLE_CENSUS_TOPIC_FILTER_OPTION | false                        | Enable filtering on various census topics
| GRACEFUL_SHUTDOWN_TIMEOUT         | 5s                           | The graceful shutdown timeout in seconds (`time.Duration` format)
| HEALTHCHECK_CRITICAL_TIMEOUT      | 90s                          | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)
| HEALTHCHECK_INTERVAL              | 30s                          | Time between self-healthchecks (`time.Duration` format)
| NO_INDEX_ENABLED                  | false                        | If true then prevents most search engine web crawlers from indexing the search pages
| PATTERN_LIBRARY_ASSETS_PATH       | ""                           | Pattern library location
| SITE_DOMAIN                       | localhost                    |
| SUPPORTED_LANGUAGES               | [2]string{"en", "cy"}        | Supported languages

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

Copyright © 2020 - 2021, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.