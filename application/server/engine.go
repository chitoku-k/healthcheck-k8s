package server

import (
	"errors"
	"fmt"
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
				logrus.Errorln("Error in health check: %s", err.Error())
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
