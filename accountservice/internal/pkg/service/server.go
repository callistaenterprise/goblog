package service

import (
	"github.com/callistaenterprise/goblog/accountservice/cmd"
	"github.com/callistaenterprise/goblog/common/monitoring"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	gqlhandler "github.com/graphql-go/graphql-go-handler"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	cfg *cmd.Config
	r   *chi.Mux
	h   *Handler
}

func NewServer(cfg *cmd.Config, h *Handler, qlResolvers GraphQLResolvers) *Server {
	initGraphQL(qlResolvers)
	return &Server{cfg: cfg, h: h}
}

func (s *Server) Close() {

}

func (s *Server) Start() {

	logrus.Infof("Starting HTTP server on %v", ":"+s.cfg.Port)
	err := http.ListenAndServe(":"+s.cfg.Port, s.r)
	if err != nil {
		logrus.WithError(err).Fatal("error starting HTTP server")
	}
}

func (s *Server) SetupRoutes() {

	s.r = chi.NewRouter()
	s.r.Use(middleware.RequestID)
	s.r.Use(middleware.RealIP)
	s.r.Use(middleware.Logger)
	s.r.Use(middleware.Recoverer)
	s.r.Use(middleware.Timeout(time.Minute))
	s.r.Use(Tracing)

	// Sub-routers with monitoring
	s.r.Route("/accounts", func(r chi.Router) {
		r.With(Monitor(s.cfg.Name, "GetAccount", "GET /accounts/{accountId}")).
			Get("/{accountId}", s.h.GetAccount)
		r.With(Monitor(s.cfg.Name, "StoreAccount", "POST /accounts")).
			Post("/", s.h.StoreAccount)
		r.With(Monitor(s.cfg.Name, "UpdateAccount", "PUT /accounts")).
			Put("/", s.h.UpdateAccount)
	})

	s.r.With(Monitor(s.cfg.Name, "GraphQL", "POST /graphql")).
		Post("/graphql", gqlhandler.New(&gqlhandler.Config{
			Schema: &schema,
			Pretty: false,
		}).ServeHTTP)

	s.r.Get("/health", s.h.HealthCheck)
	s.r.Get("/metrics", promhttp.Handler().ServeHTTP)

	logrus.Info("Successfully set up chi routes")
}

func Monitor(serviceName, routeName, signature string) func(http.Handler) http.Handler {
	summaryVec := monitoring.BuildSummaryVec(serviceName, routeName, signature)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			start := time.Now()
			next.ServeHTTP(rw, req)
			duration := time.Since(start)

			// Store duration of request
			summaryVec.WithLabelValues("duration").Observe(duration.Seconds())

			// Store size of response, if possible.
			size, err := strconv.Atoi(rw.Header().Get("Content-Length"))
			if err == nil {
				summaryVec.WithLabelValues("size").Observe(float64(size))
			}
		})
	}
}

func Tracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		span := tracing.StartHTTPTrace(req, req.RequestURI)
		defer span.Finish()

		ctx := tracing.UpdateContext(req.Context(), span)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}
