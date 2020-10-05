module github.com/callistaenterprise/goblog/dataservice

go 1.14

replace github.com/Sirupsen/logrus v1.4.2 => github.com/sirupsen/logrus v1.7.0

require (
	github.com/alexflint/go-arg v1.3.0
	github.com/callistaenterprise/goblog/common v0.0.0-20190723162557-085a94bc23ae
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/golang/mock v1.2.0
	github.com/jinzhu/gorm v1.9.16
	github.com/opentracing/opentracing-go v1.1.0
	github.com/prometheus/client_golang v1.7.1
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/twinj/uuid v1.0.0
	golang.org/x/sys v0.0.0-20201005172224-997123666555 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)
