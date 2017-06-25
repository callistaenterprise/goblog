package tracing

import (
        "net/http"
        "github.com/opentracing/opentracing-go"
        "github.com/opentracing/opentracing-go/ext"
        zipkin "github.com/openzipkin/zipkin-go-opentracing"
        "fmt"
        "github.com/Sirupsen/logrus"
)

var Tracer opentracing.Tracer

func InitTracing(zipkinUrl string, serviceName string) {
        logrus.Infof("Connecting to zipkin server at %v", zipkinUrl)
        collector, err := zipkin.NewHTTPCollector(
                fmt.Sprintf("%s/api/v1/spans", zipkinUrl))
        if err != nil {
                logrus.Info("Error connecting to zipkin server at " +
                        fmt.Sprintf("%s/api/v1/spans", zipkinUrl) + ". Error: " + err.Error())
                logrus.Errorln("Error connecting to zipkin server at " +
                        fmt.Sprintf("%s/api/v1/spans", zipkinUrl) + ". Error: " + err.Error())
                panic("Error connecting to zipkin server at " +
                        fmt.Sprintf("%s/api/v1/spans", zipkinUrl) + ". Error: " + err.Error())
        }
        Tracer, err = zipkin.NewTracer(
                zipkin.NewRecorder(collector, false, "127.0.0.1:0", serviceName))
        if err != nil {
                logrus.Errorln("Error starting new zipkin tracer. Error: " + err.Error())
                panic("Error starting new zipkin tracer. Error: " + err.Error())
        }
        logrus.Info("Successfully started zipkin tracer")
}

// Loads tracing information from an INCOMING HTTP request.
func StartHTTPTrace(r *http.Request, opName string) opentracing.Span {
        carrier := opentracing.HTTPHeadersCarrier(r.Header)
        clientContext, err := Tracer.Extract(opentracing.HTTPHeaders, carrier)
        var span opentracing.Span
        if err == nil {
                span = Tracer.StartSpan(
                        opName, ext.RPCServerOption(clientContext))
        } else {
                span = Tracer.StartSpan(opName)
        }
        return span
}

// Adds tracing information to an OUTGOING HTTP request
func AddTracingToReq(req *http.Request, span opentracing.Span) {
        carrier := opentracing.HTTPHeadersCarrier(req.Header)
        err := Tracer.Inject(
                span.Context(),
                opentracing.HTTPHeaders,
                carrier)
        if err != nil {
                panic("Unable to inject tracing context: " + err.Error())
        }
}
