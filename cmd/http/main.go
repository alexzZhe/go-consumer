package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/gops/agent"
	_ "go.uber.org/automaxprocs"
)

func main() {
	if err := agent.Listen(agent.Options{
		// Addr: "0.0.0.0:8082", // 你想监听的地址，不填的话系统会自动分配一个端口给它用。
	}); err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.POST("/test", func(c *gin.Context) {
		// time.Sleep(15 * time.Second)
		c.JSON(http.StatusOK, gin.H{
			"message": "test",
		})
	})
	r.POST("/test2", func(c *gin.Context) {
		// time.Sleep(1 * time.Second)
		c.JSON(http.StatusOK, gin.H{
			"message": "test2222",
		})
	})
	r.POST("/test3", func(c *gin.Context) {
		// time.Sleep(1 * time.Second)
		c.JSON(http.StatusOK, gin.H{
			"message": "test3333",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
