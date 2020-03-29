package main
import "github.com/gin-gonic/gin"
func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(200, "hello world")
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}