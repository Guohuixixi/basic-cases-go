package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()

	server.Use(func(ctx *gin.Context) {
		started := time.Now()
		// Next 会先执行后面的中间件和处理函数，回来后正好统计整次请求耗时。
		ctx.Next()
		println("请求耗时:", time.Since(started).String())
	})

	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello, world")
	})

	server.GET("/users/:name", func(context *gin.Context) {
		name := context.Param("name")
		context.String(http.StatusOK, "这是你传过来的名字 %s", name)
	})

	server.GET("/order", func(context *gin.Context) {
		// 查询参数
		id := context.Query("id")
		context.String(http.StatusOK, "你传过来的 ID 是 %s", id)
	})

	server.GET("/views/*.html", func(context *gin.Context) {
		path := context.Param(".html")
		context.String(http.StatusOK, "匹配上的值是 %s", path)
	})

	// 这种路由是不合法的
	//server.GET("/invalid/*", func(context *gin.Context) {
	//})

	// 这种路由也是不合法的
	//server.GET("/invalid/*/b", func(context *gin.Context) {
	//	//context.String(http.StatusOK, "")
	//})

	// 8080 是启动端口
	if err := server.Run(":8080"); err != nil {
		panic(err)
	}
}
