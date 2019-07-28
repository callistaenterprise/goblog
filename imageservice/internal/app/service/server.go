package service

import (
	"github.com/callistaenterprise/goblog/common/monitoring"
	"github.com/callistaenterprise/goblog/common/tracing"
	"github.com/callistaenterprise/goblog/imageservice/cmd"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	h   *Handler
	cfg *cmd.Config
	r   *chi.Mux
}

func NewServer(h *Handler, cfg *cmd.Config) *Server {
	return &Server{h: h, cfg: cfg}
}

func (s *Server) Close() {
	s.h.Close()
}

func (s *Server) Start() {

	logrus.Infof("Starting HTTP server at '%v'", ":"+s.cfg.Port)

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
	s.r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			logrus.Infof("Headers:\n%+v", req.Header)
			next.ServeHTTP(rw, req)
		})
	})
	s.r.Use(middleware.Recoverer)
	s.r.Use(middleware.Timeout(time.Minute))

	// Sub-routers with monitoring
	s.r.Route("/accounts", func(r chi.Router) {
		s.r.With(Trace("GetAccountImage")).With(Monitor(s.cfg.Name, "GetAccountImage", "GET /accounts/{accountId}")).Get("/{accountId}", s.h.GetAccountImage)
		s.r.With(Trace("CreateAccountImage")).With(Monitor(s.cfg.Name, "CreateAccountImage", "POST /accounts")).Post("/", s.h.CreateAccountImage)
		s.r.With(Trace("UpdateAccountImage")).With(Monitor(s.cfg.Name, "UpdateAccountImage", "PUT /accounts")).Put("/", s.h.UpdateAccountImage)
	})
	s.r.Route("/file", func(r chi.Router) {
		s.r.With(Trace("GetAccountImage")).With(Monitor(s.cfg.Name, "ProcessImage", "GET /file/{filename}")).Get("/{filename}", s.h.ProcessImageFromFile)
	})

	s.r.Get("/health", func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("OK"))
		rw.WriteHeader(200)
	})
	s.r.Get("/metrics", promhttp.Handler().ServeHTTP)
}

func (s *Server) SeedAccountImages() {
	err := s.h.dbClient.SeedAccountImages()
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

func Trace(opName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			logrus.Infof("starting span for %v", opName)
			span := tracing.StartHTTPTrace(req, opName)
			ctx := tracing.UpdateContext(req.Context(), span)
			next.ServeHTTP(rw, req.WithContext(ctx))

			span.Finish()
			logrus.Infof("finished span for %v", opName)
		})
	}
}
