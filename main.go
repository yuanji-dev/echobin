package main

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// @tag.name         HTTP methods
// @tag.description  Testing different HTTP verbs
// @tag.name         Auth
// @tag.description  Auth methods
// @tag.name         Status codes
// @tag.description  Generates responses with given status code
// @tag.name         Request inspection
// @tag.description  Inspect the request data
// @tag.name         Response inspection
// @tag.description  Inspect the response data like caching and headers
// @tag.name         Response formats
// @tag.description  Returns responses in different data formats
// @tag.name         Dynamic data
// @tag.description  Generates random and dynamic data
// @tag.name         Cookies
// @tag.description  Creates, reads and deletes Cookies
// @tag.name         Images
// @tag.description  Returns different image formats
// @tag.name         Redirects
// @tag.description  Returns different redirect responses
// @tag.name         Anything
// @tag.description  Returns anything that is passed to request
func newEcho() (e *echo.Echo) {
	e = echo.New()
	e.JSONSerializer = &echobinJSONSerializer{}

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Swagger docs
	e.GET("/*", swaggerHandler)
	e.GET("/swagger.json", swaggerDocHandler)
	// HTTP methods
	e.GET("/get", getMethodHandler)
	e.POST("/post", otherMethodHandler)
	e.PUT("/put", otherMethodHandler)
	e.PATCH("/patch", otherMethodHandler)
	e.DELETE("/delete", otherMethodHandler)
	// Auth
	e.GET("/basic-auth/:user/:passwd", basicAuthHandler, middleware.BasicAuth(basicAuthValidator))
	e.GET("/bearer", bearerHandler)
	// Status Codes
	e.Any("/status/:codes", statusCodesHandler)
	// Request inspection
	e.GET("/headers", requestHeadersHandler)
	e.GET("/ip", requestIPHandler)
	e.GET("/user-agent", requestUserAgentHandler)
	// Response inspection
	e.GET("/cache", cacheHandler)
	e.GET("/cache/:value", cacheDurationHandler)
	e.GET("/etag/:etag", etagHandler)
	e.GET("/response-headers", responseHeadersHandler)
	e.POST("/response-headers", responseHeadersHandler)
	// Response formats
	e.GET("/html", serveHTMLHandler)
	e.GET("/xml", serveXMLHandler)
	e.GET("/json", serveJSONHandler)
	e.GET("/robots.txt", serveRobotsTXTHandler)
	e.GET("/deny", serveDenyHandler)
	e.GET("/encoding/utf8", serveUTF8HTMLHandler)
	e.GET("/gzip", serveGzipHandler, middleware.Gzip())
	e.GET("/deflate", serveDeflateHandler, middleware.Deflate())
	e.GET("/brotli", serveBrotliHandler)
	// Dynamic data
	e.GET("/base64/:value", base64Handler)
	e.GET("/bytes/:n", generateBytesHandler)
	e.Any("/delay/:delay", delayHandler)
	e.GET("/drip", dripHandler)
	e.GET("/links/:n/:offset", linksHandler).Name = "links"
	e.GET("/range/:numbytes", rangeHandler)
	e.GET("/stream-bytes/:n", streamBytesHandler)
	e.GET("/stream/:n", streamHandler)
	e.GET("/uuid", UUIDHandler)
	// Cookies
	e.GET("/cookies", getCookiesHandler)
	e.GET("/cookies/delete", deleteCookiesHandler)
	e.GET("/cookies/set", setCookiesInQueryHandler)
	e.GET("/cookies/set/:name/:value", setCookiesInPathHandler)
	// Images
	e.GET("/image", imageHandler)
	e.GET("/image/webp", imageWebPHandler)
	e.GET("/image/svg", imageSVGHandler)
	e.GET("/image/jpeg", imageJPEGHandler)
	e.GET("/image/png", imagePNGHandler)
	// Redirects
	e.GET("/redirect-to", getRedirectToHandler)
	e.Match([]string{
		http.MethodDelete,
		http.MethodPatch,
		http.MethodPost,
		http.MethodPut,
	}, "/redirect-to", otherRedirectToHandler)
	e.GET("/redirect/:n", redirectHandler)
	e.GET("/absolute-redirect/:n", absoluteRedirectHandler)
	e.GET("/relative-redirect/:n", relativeRedirectHandler)
	// Anything
	e.Any("/anything*", anythingHandler)

	return
}

// @title          echobin
// @description    Yet another Golang port of [httpbin](https://httpbin.org/)(a HTTP request & response testing service), powered by [echo framework](https://echo.labstack.com/).
// @schemes        https http
// @contact.name   the developer
// @contact.url    https://github.com/masakichi/echobin
// @contact.email  self@gimo.me
func main() {
	e := newEcho()
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8080"
	}
	e.Logger.Fatal(e.Start(listenAddr))
}
