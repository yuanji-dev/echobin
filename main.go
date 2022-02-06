package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "github.com/masakichi/echobin/docs"
)

func newEcho() (e *echo.Echo) {
	e = echo.New()
	e.JSONSerializer = &echobinJSONSerializer{}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Swagger docs
	e.GET("/*", echoSwagger.WrapHandler)
	// HTTP methods
	e.GET("/get", getMethodHandler)
	e.POST("/post", otherMethodHandler)
	e.PUT("/put", otherMethodHandler)
	e.PATCH("/patch", otherMethodHandler)
	e.DELETE("/delete", otherMethodHandler)
	// Status Codes
	e.Any("/status/:codes", statusCodesHandler)
	// Request inspection
	e.GET("/headers", requestHeadersHandler)
	e.GET("/ip", requestIPHandler)
	e.GET("/user-agent", requestUserAgentHandler)
	// Response formats
	e.GET("/html", serveHTMLHandler)
	e.GET("/xml", serveXMLHandler)
	e.GET("/json", serveJSONHandler)
	e.GET("/robots.txt", serveRobotsTXTHandler)
	e.GET("/deny", serveDenyHandler)
	e.GET("/encoding/utf8", serveUTF8HTMLHandler)
	e.GET("/gzip", forceEncode(serveGzipHandler, "gzip"), middleware.Gzip())
	e.GET("/deflate", forceEncode(serveDeflateHandler, "deflate"), middleware.Deflate())
	// Dynamic data
	e.GET("/bytes/:n", generateBytesHandler)
	e.Any("/delay/:delay", delayHandler)
	e.GET("/drip", dripHandler)
	e.GET("/links/:n/:offset", linksHandler).Name = "links"
	e.GET("/stream/:n", streamHandler)
	e.GET("/uuid", UUIDHandler)
	// TODO: Auth
	// TODO: Response inspection
	// TODO: Cookies
	// TODO: Images
	// TODO: Redirects
	// TODO: Anything

	//e.Logger.Fatal(e.Start("127.0.0.1:1323"))
	return
}

// @title        echobin API
// @version      0.1
// @description  A simple HTTP Request & Response Service.

// @contact.name   Yuanji
// @contact.url    https://gimo.me
// @contact.email  self@gimo.me

// @license.name  MIT License
// @license.url   https://github.com/masakichi/echobin/blob/main/LICENSE

// @tag.name         HTTP methods
// @tag.description  Testing different HTTP verbs
// @tag.name         Status codes
// @tag.description  Generates responses with given status code
// @tag.name         Request inspection
// @tag.description  Inspect the request data
// @tag.name         Response formats
// @tag.description  Returns responses in different data formats
// @tag.name         Dynamic data
// @tag.description  Generates random and dynamic data
func main() {
	e := newEcho()
	e.Logger.Fatal(e.Start("127.0.0.1:1323"))
}

func forceEncode(h echo.HandlerFunc, encoding string) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Request().Header.Add(echo.HeaderAcceptEncoding, encoding)
		return h(c)
	}
}
