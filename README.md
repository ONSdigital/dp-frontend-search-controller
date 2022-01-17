# dp-frontend-search-controller

An HTTP service for the controlling of search API

## Getting started

* Run `make debug`

In order to see new search in the UI:
* run the [dp-design-system](https://github.com/ONSdigital/dp-design-system)
* set `SearchRoutesEnabled` to `true` in the config of the [dp-frontend-router](https://github.com/ONSdigital/dp-frontend-router)

## Dependencies

* No further dependencies other than those defined in `go.mod`

## Configuration

| Environment variable           | Default                      | Description
| ------------------------------ | -----------------------      | -----------
| BIND_ADDR                      | localhost:25000              | The host and port to bind to
| API_ROUTER_URL                 | http://localhost:23200/v1    | The URL of dp-api-router
| DEBUG                          | false                        | Enable debug mode
| API_ROUTER_URL                 | http://localhost:23200/v1    | The URL of the [dp-api-router](https://github.com/ONSdigital/dp-api-router)
| SITE_DOMAIN                    | localhost                    |
| PATTERN_LIBRARY_ASSETS_PATH    | ""                           | Pattern library location
| SUPPORTED_LANGUAGES            | [2]string{"en", "cy"}        | Supported languages
| GRACEFUL_SHUTDOWN_TIMEOUT      | 5s                           | The graceful shutdown timeout in seconds (`time.Duration` format)
| HEALTHCHECK_INTERVAL           | 30s                          | Time between self-healthchecks (`time.Duration` format)
| HEALTHCHECK_CRITICAL_TIMEOUT   | 90s                          | Time to wait until an unhealthy dependent propagates its state to make this app unhealthy (`time.Duration` format)
| DEFAULT_SORT                   | relevance                    | The default sort of search results
| DEFAULT_OFFSET                 | 0                            | The default offset of search results
| DEFAULT_PAGE                   | 1                            | The default current page of search results
| DEFAULT_LIMIT                  | 10                           | The default limit of search results in a page
| DEFAULT_MAXIMUM_LIMIT          | 50                           | The default maximum limit of search results in a page
| DEFAULT_MAXIMUM_SEARCH_RESULTS | 500                          | The default maximum search results

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details.

## License

Copyright © 2020 - 2021, Office for National Statistics (https://www.ons.gov.uk)

Released under MIT license, see [LICENSE](LICENSE.md) for details.