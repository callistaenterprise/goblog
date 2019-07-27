module github.com/callistaenterprise/goblog/dataservice

go 1.12

replace github.com/Sirupsen/logrus v1.4.2 => github.com/sirupsen/logrus v1.4.2

require (
	github.com/alexflint/go-arg v1.0.0
	github.com/callistaenterprise/goblog/common v0.0.0-20190723162557-085a94bc23ae
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/golang/mock v1.2.0
	github.com/jinzhu/gorm v1.9.10
	github.com/opentracing/opentracing-go v1.1.0
	github.com/prometheus/client_golang v1.0.0
	github.com/sirupsen/logrus v1.2.0
	github.com/stretchr/testify v1.3.0
	github.com/twinj/uuid v1.0.0
)
