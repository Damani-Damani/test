package webserver

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	ws "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"controlserver/auth"
	rt "controlserver/realtime"
)

var (
	upgrader = ws.Upgrader{}
)

type ControlServerContext struct {
	echo.Context
	g  context.Context
	cm *rt.ConnectionManager
}

func WsHandler(c echo.Context) error {
	ctx := c.(*ControlServerContext)
	robotIdStr := c.Param("robotId")
	robotId, err := strconv.Atoi(robotIdStr)
	if err != nil {
		return echo.ErrBadRequest
	}

	if userToken := c.Request().Header.Get("Authorization"); !auth.CanUserConnectToRobotID(userToken, robotId) {
		return echo.ErrForbidden
	}

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return echo.ErrInternalServerError
	}
	defer ws.Close()

	var client *Client
	client = &Client{
		cm:      ctx.cm,
		conn:    ws,
		robotId: robotId,
		Send:    make(chan []byte),
	}
	ctx.cm.RegisterClient(robotId, client)

	client.conn.SetCloseHandler(func(code int, text string) error {
		client.cm.RobotConnections[robotId].RemoveClient(client)
		return nil
	})

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		client.Sender()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		client.Reader()
	}()

	wg.Wait()
	return nil
}

func CreateServer(ctx context.Context, cm *rt.ConnectionManager) {
	e := echo.New()

	go func() {
		<-ctx.Done()
		ctx2, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx2); err != nil {
			e.Logger.Fatal(err)
		}
	}()

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &ControlServerContext{
				Context: c,
				g:       ctx,
				cm:      cm,
			}
			return next(cc)
		}
	})
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	authenticatedRoutes := e.Group("/ws")
	authenticatedRoutes.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if userToken := c.Request().Header.Get("Authorization"); !auth.IsUserAuthenticated(userToken) {
				return echo.ErrForbidden
			}
			return next(c)
		}
	})

	authenticatedRoutes.GET("/robot/:robotId", WsHandler)
	e.Logger.Fatal(e.Start(":8080"))
}
