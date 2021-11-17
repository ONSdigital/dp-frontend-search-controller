package data

import (
	"context"
	"net/url"
	"strconv"
	"testing"

	"github.com/ONSdigital/dp-frontend-search-controller/config"
	"github.com/ONSdigital/dp-renderer/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitReviewPaginationSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given valid limit and page", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		urlQuery := url.Values{
			"limit": []string{"10"},
			"page":  []string{"1"},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewPagination is called", func() {
			err := reviewPagination(ctx, cfg, urlQuery, validatedQueryParams)

			Convey("Then return no error", func() {
				So(err, ShouldBeNil)
			})

			Convey("And successfully set pagination parameters in validatedQueryParams", func() {
				So(validatedQueryParams.Limit, ShouldEqual, 10)
				So(validatedQueryParams.CurrentPage, ShouldEqual, 1)
				So(validatedQueryParams.Offset, ShouldEqual, 0)
			})
		})
	})
}

func TestUnitReviewPaginationFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given the failure to get offset due to invalid page given", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		urlQuery := url.Values{
			"limit": []string{"10"},
			"page":  []string{"1000000"},
		}

		validatedQueryParams := &SearchURLParams{}

		Convey("When reviewPagination is called", func() {
			err := reviewPagination(ctx, cfg, urlQuery, validatedQueryParams)

			Convey("Then return error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestUnitGetLimitFromURLQuerySuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given a limit between DefaultLimit and DefaultMaximumLimit", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		query := url.Values{
			"limit": []string{"20"},
		}

		Convey("When getLimitFromURLQuery is called", func() {
			limit := getLimitFromURLQuery(ctx, cfg, query)

			Convey("Then successfully return the limit as integer", func() {
				So(limit, ShouldEqual, 20)
			})
		})
	})

	Convey("Given a limit that is not an integer", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		query := url.Values{
			"limit": []string{"INVALID"},
		}

		Convey("When getLimitFromURLQuery is called", func() {
			limit := getLimitFromURLQuery(ctx, cfg, query)

			Convey("Then successfully return the DefaultLimit as integer", func() {
				So(limit, ShouldEqual, cfg.DefaultLimit)
			})

			Convey("And set query's limit parameter to DefaultLimit value ", func() {
				So(query.Get("limit"), ShouldEqual, strconv.Itoa(cfg.DefaultLimit))
			})
		})
	})

	Convey("Given a limit less than DefaultLimit", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		query := url.Values{
			"limit": []string{"-2"},
		}

		Convey("When getLimitFromURLQuery is called", func() {
			limit := getLimitFromURLQuery(ctx, cfg, query)

			Convey("Then successfully return the DefaultLimit as integer", func() {
				So(limit, ShouldEqual, cfg.DefaultLimit)
			})

			Convey("And set query's limit parameter to DefaultLimit value ", func() {
				So(query.Get("limit"), ShouldEqual, strconv.Itoa(cfg.DefaultLimit))
			})
		})
	})

	Convey("Given a limit more than DefaultMaximumLimit", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		query := url.Values{
			"limit": []string{"100000000"},
		}

		Convey("When getLimitFromURLQuery is called", func() {
			limit := getLimitFromURLQuery(ctx, cfg, query)

			Convey("Then successfully return the DefaultLimit as integer", func() {
				So(limit, ShouldEqual, cfg.DefaultMaximumLimit)
			})

			Convey("And set query's limit parameter to defaultMaximumLimit value ", func() {
				So(query.Get("limit"), ShouldEqual, strconv.Itoa(cfg.DefaultMaximumLimit))
			})
		})
	})
}

func TestUnitGetPageFromURLQuerySuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given a valid page", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		query := url.Values{
			"page": []string{"2"},
		}

		Convey("When getPageFromURLQuery is called", func() {
			page := getPageFromURLQuery(ctx, cfg, query)

			Convey("Then successfully return page as integer", func() {
				So(page, ShouldEqual, 2)
			})
		})
	})

	Convey("Given an invalid page which is not an integer", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		query := url.Values{
			"page": []string{"invalid"},
		}

		Convey("When getPageFromURLQuery is called", func() {
			page := getPageFromURLQuery(ctx, cfg, query)

			Convey("Then successfully return DefaultPage as integer", func() {
				So(page, ShouldEqual, cfg.DefaultPage)
			})

			Convey("And set query's page parameter to DefaultPage value", func() {
				So(query.Get("page"), ShouldEqual, strconv.Itoa(cfg.DefaultPage))
			})
		})
	})

	Convey("Given an invalid page which is less than 1", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		query := url.Values{
			"page": []string{"-2"},
		}

		Convey("When getPageFromURLQuery is called", func() {
			page := getPageFromURLQuery(ctx, cfg, query)

			Convey("Then successfully return DefaultPage as integer", func() {
				So(page, ShouldEqual, cfg.DefaultPage)
			})

			Convey("And set query's page parameter to DefaultPage value", func() {
				So(query.Get("page"), ShouldEqual, strconv.Itoa(cfg.DefaultPage))
			})
		})
	})
}

func TestUnitGetOffsetSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given valid page and/or limit given", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		page := 1
		limit := 10

		Convey("When getOffset is called", func() {
			offset, err := getOffset(ctx, cfg, page, limit)

			Convey("Then successfully get offset", func() {
				So(offset, ShouldEqual, 0)
			})

			Convey("And return no error", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given negative current page number or limit", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		page := -1
		limit := 10

		Convey("When getOffset is called", func() {
			offset, err := getOffset(ctx, cfg, page, limit)

			Convey("Then successfully get DefaultOffset", func() {
				So(offset, ShouldEqual, cfg.DefaultOffset)
			})

			Convey("And return no error", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestUnitGetOffsetFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	Convey("Given a large current page number and/or limit", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		page := 10000000
		limit := 10

		Convey("When getOffset is called", func() {
			offset, err := getOffset(ctx, cfg, page, limit)

			Convey("Then return error", func() {
				So(err, ShouldNotBeNil)
			})

			Convey("And set offset to default", func() {
				So(offset, ShouldEqual, cfg.DefaultOffset)
			})
		})
	})
}

func TestUnitGetTotalPagesSuccess(t *testing.T) {
	t.Parallel()
	cfg, _ := config.Get()

	Convey("Given valid limit and/or count", t, func() {
		limit := 10
		count := 100

		Convey("When GetTotalPages is called", func() {
			totalPages := GetTotalPages(cfg, limit, count)

			Convey("Then successfully get total pages", func() {
				So(totalPages, ShouldEqual, 10)
			})
		})

		Convey("When results count is greater than default max results GetTotalPages returns the max default", func() {
			largerCount := cfg.DefaultMaximumSearchResults + 1
			totalPages := GetTotalPages(cfg, limit, largerCount)
			expectedNumberOfPages := cfg.DefaultMaximumSearchResults / limit
			So(totalPages, ShouldEqual, expectedNumberOfPages)
		})
	})
}

func TestUnitGetPagesToDisplaySuccess(t *testing.T) {
	t.Parallel()

	Convey("Given validated query parameters and total pages", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		validatedQueryParams := SearchURLParams{
			Query: "housing",
			Filter: Filter{
				Query:           []string{"article"},
				LocaliseKeyName: []string{"Article"},
			},
			Sort: Sort{
				Query:           "relevance",
				LocaliseKeyName: "Relevance",
			},
			Limit:       10,
			CurrentPage: 1,
			Offset:      0,
		}

		totalPages := 5

		Convey("When GetPagesToDisplay is called", func() {
			pagesToDisplay := GetPagesToDisplay(cfg, validatedQueryParams, totalPages)

			Convey("Then return all available page numbers with its respective URL", func() {
				So(pagesToDisplay, ShouldResemble, []model.PageToDisplay{
					{
						PageNumber: 1,
						URL:        "/search?q=housing&filter=article&limit=10&sort=relevance&page=1",
					},
					{
						PageNumber: 2,
						URL:        "/search?q=housing&filter=article&limit=10&sort=relevance&page=2",
					},
					{
						PageNumber: 3,
						URL:        "/search?q=housing&filter=article&limit=10&sort=relevance&page=3",
					},
					{
						PageNumber: 4,
						URL:        "/search?q=housing&filter=article&limit=10&sort=relevance&page=4",
					},
					{
						PageNumber: 5,
						URL:        "/search?q=housing&filter=article&limit=10&sort=relevance&page=5",
					},
				})
			})
		})
	})
}

func TestUnitGetStartPageSuccess(t *testing.T) {
	t.Parallel()

	Convey("Given currentPage is between 3 and 3rd last page", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		currentPage := 10
		totalPages := 20

		Convey("When getStartPage is called", func() {
			startPage := getStartPage(cfg, currentPage, totalPages)

			Convey("Then successfully return start page", func() {
				So(startPage, ShouldEqual, 8)
			})
		})
	})

	Convey("Given current page is less than or equal to 2", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		currentPage := 2
		totalPages := 20

		Convey("When getStartPage is called", func() {
			startPage := getStartPage(cfg, currentPage, totalPages)

			Convey("Then successfully return first page", func() {
				So(startPage, ShouldEqual, 1)
			})
		})
	})

	Convey("Given current page is second last page", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		currentPage := 19
		totalPages := 20

		Convey("When getStartPage is called", func() {
			startPage := getStartPage(cfg, currentPage, totalPages)

			Convey("Then successfully return fifth last page", func() {
				So(startPage, ShouldEqual, 16)
			})
		})
	})

	Convey("Given current page is last page", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		currentPage := 20
		totalPages := 20

		Convey("When getStartPage is called", func() {
			startPage := getStartPage(cfg, currentPage, totalPages)

			Convey("Then successfully return fifth last page", func() {
				So(startPage, ShouldEqual, 16)
			})
		})
	})
}

func TestUnitGetEndPageSuccess(t *testing.T) {
	t.Parallel()

	Convey("Given valid start page", t, func() {
		startPage := 10

		Convey("And total pages is more than or equal to 5", func() {
			totalPages := 20

			Convey("When getEndPage is called", func() {
				endPage := getEndPage(startPage, totalPages)

				Convey("Then successfully return end page", func() {
					So(endPage, ShouldEqual, 14)
				})
			})
		})
	})

	Convey("Given valid start page", t, func() {
		startPage := 3

		Convey("And total pages is less than 5", func() {
			totalPages := 4

			Convey("When getEndPage is called", func() {
				endPage := getEndPage(startPage, totalPages)

				Convey("Then successfully return end page", func() {
					So(endPage, ShouldEqual, 4)
				})
			})
		})
	})
}

func TestUnitGetPageURLSuccess(t *testing.T) {
	t.Parallel()

	Convey("Given search query, page and controller query", t, func() {
		query := "housing"
		page := 1

		controllerQuery := url.Values{
			"q":      []string{"housing"},
			"filter": []string{"article"},
			"sort":   []string{"relevance"},
			"limit":  []string{"10"},
			"page":   []string{"1"},
		}

		Convey("When getPageURL is called", func() {
			pageURL := getPageURL(query, page, controllerQuery)

			Convey("Then successfully return page URL with query first and page last", func() {
				So(pageURL, ShouldEqual, "/search?q=housing&filter=article&limit=10&sort=relevance&page=1")
			})
		})
	})

	Convey("Given no filter, sort, limit in controllerQuery", t, func() {
		query := "housing"
		page := 1

		controllerQuery := url.Values{
			"q":    []string{"housing"},
			"page": []string{"1"},
		}

		Convey("When getPageURL is called", func() {
			pageURL := getPageURL(query, page, controllerQuery)

			Convey("Then successfully return page URL with query and page", func() {
				So(pageURL, ShouldEqual, "/search?q=housing&page=1")
			})
		})
	})
}
