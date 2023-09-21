module github.com/ONSdigital/dp-frontend-search-controller

go 1.21

// to fix: [CVE-2023-32731] CWE-Other
replace google.golang.org/grpc => google.golang.org/grpc v1.55.0

// to avoid the following vulnerabilities:
//     - CVE-2022-29153 # pkg:golang/github.com/hashicorp/consul/api@v1.1.0 and pkg:golang/github.com/hashicorp/consul/sdk@v0.1.1
//     - sonatype-2021-1401 # pkg:golang/github.com/miekg/dns@v1.0.14
//     - sonatype-2019-0890 # pkg:golang/github.com/pkg/sftp@v1.10.1
replace github.com/spf13/cobra => github.com/spf13/cobra v1.4.0

require (
	github.com/ONSdigital/dp-api-clients-go/v2 v2.254.0
	github.com/ONSdigital/dp-cache v0.1.0
	github.com/ONSdigital/dp-component-test v0.9.2
	github.com/ONSdigital/dp-cookies v0.3.3
	github.com/ONSdigital/dp-healthcheck v1.6.1
	github.com/ONSdigital/dp-net/v2 v2.10.0
	github.com/ONSdigital/dp-renderer/v2 v2.1.1
	github.com/ONSdigital/dp-search-api v1.38.0
	github.com/ONSdigital/dp-topic-api v0.13.3
	github.com/ONSdigital/log.go/v2 v2.4.1
	github.com/cucumber/godog v0.12.6
	github.com/gorilla/mux v1.8.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/kevinburke/go-bindata v3.24.0+incompatible
	github.com/maxcnunes/httpfake v1.2.4
	github.com/smartystreets/goconvey v1.8.0
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/ONSdigital/dp-authorisation v0.2.1 // indirect
	github.com/ONSdigital/dp-elasticsearch/v3 v3.0.1-alpha.4.0.20230308115225-bb7559a89d0c // indirect
	github.com/ONSdigital/dp-mongodb-in-memory v1.7.0 // indirect
	github.com/ONSdigital/dp-rchttp v1.0.0 // indirect
	github.com/ONSdigital/go-ns v0.0.0-20210916104633-ac1c1c52327e // indirect
	github.com/c2h5oh/datasize v0.0.0-20220606134207-859f65c6625b // indirect
	github.com/chromedp/cdproto v0.0.0-20230722233645-dbf72f61037f // indirect
	github.com/chromedp/chromedp v0.9.1 // indirect
	github.com/chromedp/sysutil v1.0.0 // indirect
	github.com/cucumber/gherkin-go/v19 v19.0.3 // indirect
	github.com/cucumber/messages-go/v16 v16.0.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/elastic/go-elasticsearch/v7 v7.17.10 // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.2.1 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/gosimple/slug v1.13.1 // indirect
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-memdb v1.3.4 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/justinas/alice v1.2.0 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/nicksnyder/go-i18n/v2 v2.2.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/smartystreets/assertions v1.13.1 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/tdewolff/minify v2.3.6+incompatible // indirect
	github.com/tdewolff/parse v2.3.4+incompatible // indirect
	github.com/unrolled/render v1.6.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	go.mongodb.org/mongo-driver v1.12.0 // indirect
	golang.org/x/crypto v0.11.0 // indirect
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
