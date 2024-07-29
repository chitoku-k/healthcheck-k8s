package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/chitoku-k/healthcheck-k8s/service"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

type engine struct {
	Port           string
	HeaderName     string
	TrustedProxies []string
	HealthCheck    service.HealthCheck
}

type Engine interface {
	Start(ctx context.Context) error
}

func NewEngine(
	port string,
	headerName string,
	trustedProxies []string,
	healthCheck service.HealthCheck,
) Engine {
	return &engine{
		Port:           port,
		HeaderName:     textproto.CanonicalMIMEHeaderKey(headerName),
		TrustedProxies: trustedProxies,
		HealthCheck:    healthCheck,
	}
}

func (e *engine) Start(ctx context.Context) error {
	router := gin.New()
	err := router.SetTrustedProxies(e.TrustedProxies)
	if err != nil {
		return fmt.Errorf("invalid proxies: %w", err)
	}

	router.Use(gin.Recovery())
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: e.Formatter(),
		SkipPaths: []string{"/healthz"},
	}))

	router.Any("/", func(c *gin.Context) {
		h := c.Request.Header.Values(e.HeaderName)
		if len(h) == 0 {
			c.String(http.StatusBadRequest, fmt.Sprintf(`Header "%s" was not specified.`, e.HeaderName))
			return
		}

		for _, node := range h {
			res, err := e.HealthCheck.Do(c, node)
			if service.IsNotFound(err) {
				c.String(http.StatusNotFound, fmt.Sprintf(`Node "%s" was not found.`, node))
				return
			}
			if service.IsTimeout(err) {
				slog.Error("Timeout in health check", slog.Any("err", err))
				c.String(http.StatusGatewayTimeout, fmt.Sprintf(`Timed out while processing node "%s".`, node))
				return
			}
			if err != nil {
				slog.Error("Error in health check", slog.Any("err", err))
				c.String(http.StatusInternalServerError, fmt.Sprintf(`Internal server error while processing node "%s".`, node))
				return
			}

			if !res {
				c.String(http.StatusServiceUnavailable, fmt.Sprintf(`Node "%s" is currently undergoing maintenance.`, node))
				return
			}
		}

		c.String(http.StatusOK, fmt.Sprintf(`Node(s) are OK: "%s"`, strings.Join(h, `", "`)))
	})

	router.Any("/ip", func(c *gin.Context) {
		c.String(http.StatusOK, c.ClientIP())
	})

	router.Any("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	server := http.Server{
		Addr:    net.JoinHostPort("", e.Port),
		Handler: router,
	}

	var eg errgroup.Group
	eg.Go(func() error {
		<-ctx.Done()
		return server.Shutdown(context.Background())
	})

	err = server.ListenAndServe()
	if err == http.ErrServerClosed {
		return eg.Wait()
	}

	return err
}

func (e *engine) Formatter() gin.LogFormatter {
	return func(param gin.LogFormatterParams) string {
		remoteHost, _, err := net.SplitHostPort(param.Request.RemoteAddr)
		if remoteHost == "" || err != nil {
			remoteHost = "-"
		}

		bodySize := fmt.Sprintf("%v", param.BodySize)
		if param.BodySize == 0 {
			bodySize = "-"
		}

		referer := param.Request.Header.Get("Referer")
		if referer == "" {
			referer = "-"
		}

		userAgent := param.Request.Header.Get("User-Agent")
		if userAgent == "" {
			userAgent = "-"
		}

		forwardedFor := param.Request.Header.Get("X-Forwarded-For")
		if forwardedFor == "" {
			forwardedFor = "-"
		}

		nodeName := param.Request.Header.Get(e.HeaderName)
		if nodeName == "" {
			nodeName = "-"
		}

		return fmt.Sprintf(`%s %s %s [%s] "%s %s %s" %v %s "%s" "%s" "%s" "%s"%s`,
			remoteHost,
			"-",
			"-",
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Request.Method,
			param.Request.RequestURI,
			param.Request.Proto,
			param.StatusCode,
			bodySize,
			referer,
			userAgent,
			forwardedFor,
			nodeName,
			"\n",
		)
	}
}
