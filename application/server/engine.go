package server

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/chitoku-k/healthcheck-k8s/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type engine struct {
	Port        string
	HeaderName  string
	HealthCheck service.HealthCheck
}

type Engine interface {
	Start()
}

func NewEngine(
	port string,
	headerName string,
	healthCheck service.HealthCheck,
) Engine {
	return &engine{
		Port:        port,
		HeaderName:  headerName,
		HealthCheck: healthCheck,
	}
}

func (e *engine) Start() {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: e.Formatter(),
		SkipPaths: []string{"/healthz"},
	}))

	router.Any("/", func(c *gin.Context) {
		h, ok := c.Request.Header[e.HeaderName]
		if !ok {
			c.String(http.StatusBadRequest, fmt.Sprintf(`Header "%s" was not specified.`, e.HeaderName))
			return
		}

		for _, node := range h {
			res, err := e.HealthCheck.Do(c, node)
			if errors.Is(err, service.ErrNotFound) {
				c.String(http.StatusNotFound, fmt.Sprintf(`Node "%s" was not found.`, node))
				return
			}
			if err != nil {
				logrus.Errorln("Error in health check:", err.Error())
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

	router.Any("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	router.Run(":" + e.Port)
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
