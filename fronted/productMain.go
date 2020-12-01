package main

import (
	"github.com/kataras/iris/v12"
)

//CDN
func main() {
	//1.创建iris 实例
	app := iris.New()
	//2.设置模板目录
	app.HandleDir("/public", "./fronted/web/public")
	//3.访问生成好的HTML静态文件
	app.HandleDir("/html", "./fronted/web/htmlProductShow")
	//4.启动服务
	app.Run(
		iris.Addr("localhost:80"),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithOptimizations,
	)
}
