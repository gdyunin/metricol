package delivery

import (
	"NewNewMetricol/internal/server/delivery/handle/general"
	"NewNewMetricol/internal/server/delivery/handle/update"
	"NewNewMetricol/internal/server/delivery/handle/value"
	custMiddleware "NewNewMetricol/internal/server/delivery/middleware"
	"NewNewMetricol/internal/server/delivery/render"
	"NewNewMetricol/internal/server/internal/usecase"
	"NewNewMetricol/internal/server/repository"
	"fmt"
	"html/template"
	"path"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

const DefaultTemplatesPath = "web/templates/"

type EchoServer struct {
	echo        *echo.Echo
	logger      *zap.SugaredLogger
	metricsCtrl *usecase.MetricService
	addr        string
	tmplPath    string
}

func NewEchoServer(serverAddress string, repo repository.Repository, logger *zap.SugaredLogger) *EchoServer {
	echoServer := EchoServer{
		echo:        echo.New(),
		logger:      logger,
		addr:        serverAddress,
		tmplPath:    DefaultTemplatesPath,
		metricsCtrl: usecase.NewMetricService(repo),
	}
	return echoServer.build()
}

func (s *EchoServer) Start() {
	err := s.echo.Start(s.addr)
	if err != nil {
		fmt.Printf("err: %v", err)
	}
}

func (s *EchoServer) build() *EchoServer {
	buildSteps := []func(){
		s.setupPreMiddlewares,
		s.setupGeneralMiddlewares,
		s.setupRenderers,
		s.setupRouters,
	}

	for _, stepFunc := range buildSteps {
		stepFunc()
	}

	return s
}

func (s *EchoServer) setupPreMiddlewares() {
	s.echo.Pre(
		echoMiddleware.RequestID(),
		echoMiddleware.RemoveTrailingSlash(),
	)
}

func (s *EchoServer) setupGeneralMiddlewares() {
	s.echo.Use(
		custMiddleware.Log(s.logger.Named("request")),
		echoMiddleware.Decompress(),
		echoMiddleware.Gzip(),
	)
}

func (s *EchoServer) setupRenderers() {
	s.echo.Renderer = render.NewRenderer(
		template.Must(template.ParseGlob(path.Join(s.tmplPath, "*.html"))),
	)
}

func (s *EchoServer) setupRouters() {
	updateGroup := s.echo.Group("/update")
	updateGroup.POST("", update.FromJSON(s.metricsCtrl))
	updateGroup.POST("/:type/:id/:value", update.FromURI(s.metricsCtrl))

	valueGroup := s.echo.Group("/value")
	valueGroup.POST("", value.FromJSON(s.metricsCtrl))
	valueGroup.GET("/:type/:id", value.FromURI(s.metricsCtrl))

	s.echo.GET("/", general.MainPage(s.metricsCtrl))
	s.echo.GET("/ping", general.Ping())

}
