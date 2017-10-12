package tracing

import (
        "net/http"
        "github.com/opentracing/opentracing-go"
        "github.com/opentracing/opentracing-go/ext"
        zipkin "github.com/openzipkin/zipkin-go-opentracing"
        "fmt"
        "github.com/Sirupsen/logrus"
        "context"
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
        logrus.Infof("Successfully started zipkin tracer for service '%v'", serviceName)
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

// Converts a generic map to opentracing http headers carrier
func MapToCarrier(headers map[string]interface{}) opentracing.HTTPHeadersCarrier {
        carrier := make(opentracing.HTTPHeadersCarrier)
        for k, v := range headers {    // delivery.Headers
                carrier.Set(k, v.(string))
        }
        return carrier
}

// Converts a TextMapCarrier to the amqp headers format
func CarrierToMap(values map[string]string) map[string]interface{} {
        headers := make(map[string]interface{})
        for k, v := range values {
                headers[k] = v
        }
        return headers
}

func StartTraceFromCarrier(carrier map[string]interface{}, spanName string) opentracing.Span {

        clientContext, err := Tracer.Extract(opentracing.HTTPHeaders, MapToCarrier(carrier))
        var span opentracing.Span
        if err == nil {
                span = Tracer.StartSpan(
                        spanName, ext.RPCServerOption(clientContext))
        } else {
                span = Tracer.StartSpan(spanName)
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

// Adds tracing information to an OUTGOING HTTP request
func AddTracingToReqFromContext(ctx context.Context, req *http.Request) {
        if ctx.Value("opentracing-span") ==   nil {
                return
        }
        carrier := opentracing.HTTPHeadersCarrier(req.Header)
        err := Tracer.Inject(
                ctx.Value("opentracing-span").(opentracing.Span).Context(),
                opentracing.HTTPHeaders,
                carrier)
        if err != nil {
                panic("Unable to inject tracing context: " + err.Error())
        }
}

func StartSpanFromContext(ctx context.Context, opName string) opentracing.Span {
        span := ctx.Value("opentracing-span").(opentracing.Span)
        child := Tracer.StartSpan(opName, ext.RPCServerOption(span.Context()))
        return child
}

// Starts a child span from span within the supplied context, if available.
func StartChildSpanFromContext(ctx context.Context, opName string) opentracing.Span {
        if ctx.Value("opentracing-span") == nil {
                return Tracer.StartSpan(opName, ext.RPCServerOption(nil))
        }
        parent := ctx.Value("opentracing-span").(opentracing.Span)
        child := Tracer.StartSpan(opName, opentracing.ChildOf(parent.Context()))
        return child
}
func StartSpanFromContextWithLogEvent(ctx context.Context, opName string, logStatement string) opentracing.Span {
        span := ctx.Value("opentracing-span").(opentracing.Span)
        child := Tracer.StartSpan(opName, ext.RPCServerOption(span.Context()))
        child.LogEvent(logStatement)
        return child
}

func CloseSpan(span opentracing.Span, event string) {
        span.LogEvent(event)
        span.Finish()
}

func LogEventToOngoingSpan(ctx context.Context, logMessage string) {
        if ctx.Value("opentracing-span") != nil {
                ctx.Value("opentracing-span").(opentracing.Span).LogEvent(logMessage)
        }
}

func UpdateContext(ctx context.Context, span opentracing.Span) context.Context {
        return context.WithValue(ctx, "opentracing-span", span)
}
