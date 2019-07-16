package service

import (
	"github.com/callistaenterprise/goblog/common/monitoring"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/callistaenterprise/goblog/imageservice/cmd"
	"github.com/callistaenterprise/goblog/imageservice/internal/pkg/dbclient"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	dbClient dbclient.IGormClient
	cfg      *cmd.Config
	r        *chi.Mux
}

func NewServer(dbClient dbclient.IGormClient, cfg *cmd.Config) *Server {
	return &Server{dbClient: dbClient, cfg: cfg}
}

func (s *Server) Close() {
	s.dbClient.Close()
}

func (s *Server) Start() {

	err := http.ListenAndServe(s.cfg.Port, s.r)
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
		s.r.With(Monitor(s.cfg.Name, "GetAccountImage", "GET /accounts/{accountId}")).Get("/{accountId}", s.GetAccountImage)
		s.r.With(Monitor(s.cfg.Name, "CreateAccountImage", "POST /accounts")).Post("/", s.CreateAccountImage)
		s.r.With(Monitor(s.cfg.Name, "UpdateAccountImage", "PUT /accounts")).Put("/", s.UpdateAccountImage)
	})
	s.r.Route("/file", func(r chi.Router) {
		s.r.With(Monitor(s.cfg.Name, "ProcessImage", "GET /file/{filename}")).Get("/{filename}", s.ProcessImageFromFile)
	})

	s.r.Get("/health", func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("OK"))
		rw.WriteHeader(200)
	})
	s.r.Get("/metrics", promhttp.Handler().ServeHTTP)
}

func (s *Server) SeedAccountImages() {
	err := s.dbClient.SeedAccountImages()
	if err != nil {
		logrus.WithError(err).Fatal("error seeding account images")
	}
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
