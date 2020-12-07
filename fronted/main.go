package main

import (
	"context"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/opentracing/opentracing-go/log"
	"goseckill/common"
	"goseckill/fronted/middleware"
	"goseckill/fronted/web/controllers"
	"goseckill/rabbitmq"
	"goseckill/repositories"
	"goseckill/services"
)

func main() {
	//1.创建iris 实例
	app := iris.New()
	//2.设置错误模式，在mvc模式下提示错误
	app.Logger().SetLevel("debug")
	//3.注册模板
	template := iris.HTML("./fronted/web/views", ".html").Layout("shared/layout.html").Reload(true)
	app.RegisterView(template)
	//4.设置模板目录
	app.HandleDir("/public", "./fronted/web/public")
	//访问生成好的HTML静态文件
	app.HandleDir("/html", "./fronted/web/htmlProductShow")
	//出现异常跳转到指定页面
	app.OnAnyErrorCode(func(ctx iris.Context) {
		ctx.ViewData("message", ctx.Values().GetStringDefault("message", "访问的页面出错！"))
		ctx.ViewLayout("")
		ctx.View("shared/error.html")
	})
	//连接数据库
	db, err := common.NewMysqlConn()
	if err != nil {
		log.Error(err)
	}
	//sess := sessions.New(sessions.Config{
	//	Cookie:  "AdminCookie",
	//	Expires: 600 * time.Minute,
	//})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//5.注册控制器
	user := repositories.NewUserRepository("user", db)
	userService := services.NewService(user)
	userPro := mvc.New(app.Party("/user"))
	//userPro.Register(userService, ctx, sess.Start)
	userPro.Register(userService, ctx)
	userPro.Handle(new(controllers.UserController))

	//创建RabbitMQ实例
	rabbitmq := rabbitmq.NewRabbitMQSimple("goseckill")
	product := repositories.NewProductManagerRepository("product", db)
	productService := services.NewProductService(product)
	order := repositories.NewOrderMangerRepository("order", db)
	orderService := services.NewOrderService(order)
	proProduct := app.Party("/product")
	pro := mvc.New(proProduct)
	//权限校验
	proProduct.Use(middleware.AuthConProduct)
	pro.Register(productService, orderService, ctx, rabbitmq)
	pro.Handle(new(controllers.ProductController))
	//6.启动服务
	app.Run(
		iris.Addr("localhost:8082"),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithOptimizations,
	)
}
