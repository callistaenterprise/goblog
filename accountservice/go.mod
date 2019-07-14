module github.com/callistaenterprise/goblog/accountservice

go 1.12

replace github.com/callistaenterprise/goblog/common => ../common

require (
	github.com/callistaenterprise/goblog/common v0.0.0-20190713133714-ded5832e931e
	github.com/gorilla/mux v1.7.3
	github.com/graphql-go/graphql v0.7.8
	github.com/graphql-go/graphql-go-handler v0.2.3
	github.com/graphql-go/handler v0.2.3 // indirect
	github.com/myesui/uuid v1.0.0 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/prometheus/client_golang v1.0.0
	github.com/sirupsen/logrus v1.4.2
	github.com/smartystreets/goconvey v0.0.0-20190710185942-9d28bd7c0945
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.3.0
	gopkg.in/h2non/gock.v1 v1.0.15
)
