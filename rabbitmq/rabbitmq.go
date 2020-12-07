package rabbitmq

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"goseckill/datamodels"
	"goseckill/services"
	"log"
	"sync"
)

//url 格式 amqp://账号:密码@rabbitmq服务器地址:端口号/vhost
const MQURL = "amqp://wenjun:wenjun@114.67.94.142:5672/imooc"

type RabbitMQ struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	QueueName string
	Exchange  string
	Key       string
	MqUrl     string
	sync.Mutex
}

//创建RabbitMQ结构体实例
func NewRabbitMQ(queueName string, exchange string, key string) *RabbitMQ {
	rabbitmq := &RabbitMQ{QueueName: queueName, Exchange: exchange, Key: key, MqUrl: MQURL}
	var err error
	//创建RabbitMQ连接
	rabbitmq.conn, err = amqp.Dial(rabbitmq.MqUrl)
	rabbitmq.failOnErr(err, "创建连接错误！")
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "获取channel失败！")
	return rabbitmq
}

//断开channel和connection
func (r *RabbitMQ) Destroy() {
	r.channel.Close()
	r.conn.Close()
}

//错误处理函数
func (r *RabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s", message, err)
	}
}

//简单模式Step:1.创建简单模式下RabbitMQ实例
func NewRabbitMQSimple(queueName string) *RabbitMQ {
	return NewRabbitMQ(queueName, "", "")
}

//简单模式Step:2.简单模式下生产代码
func (r *RabbitMQ) PublishSimple(message string) error {
	//1.申请队列，如果队列不存在则自动创建，否则跳过创建
	//保证队列存在，消息发送到队列中
	r.Lock()
	defer r.Unlock()
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		//是否持久化
		false,
		//是否为自动删除
		false,
		//是否具有排他性
		false,
		//是否阻塞
		false,
		//额外属性
		nil,
	)
	if err != nil {
		return err
	}
	//2.发送消息到队列中
	err = r.channel.Publish(
		r.Exchange,
		r.QueueName,
		//如果为true，根据exchange类型和routkey规则，如果无法找到符合条件的队列，那么会把发送的消息返回给发送者
		false,
		//如果为true，当exchange发送消息到队列后发现队列上没有绑定消费者，则会把消息返回给发送者
		false,
		//发送的消息
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	return err
}

//简单模式Step:3.简单模式下消费代码
func (r *RabbitMQ) ConsumerSimple() {
	//1.申请队列，如果队列不存在则自动创建，否则跳过创建
	//保证队列存在，消息发送到队列中
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		//是否持久化
		false,
		//是否为自动删除
		false,
		//是否具有排他性
		false,
		//是否阻塞
		false,
		//额外属性
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
	msgs, err := r.channel.Consume(
		r.QueueName,
		//用来区分多个消费者
		"",
		//是否自动应答
		true,
		//是否具有排他性
		false,
		//如果设置为true，表示不能将同一个connection中发送的消息传递给这个connection中的消费者
		false,
		//队列消费是否阻塞(设置为阻塞)
		false,
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
	forever := make(chan bool)
	//启用协程处理消息
	go func() {
		for d := range msgs {
			//实现我们要处理的逻辑函数
			log.Printf("Received a message: %s", d.Body)
		}
	}()
	log.Println("[*] Waiting for messages, To exit press CTRL+C.")
	//阻塞一下
	<-forever
}

//改造后简单模式消费代码
func (r *RabbitMQ) ConsumerSimpleAlter(orderService services.IOrderService, productService services.IProductService) {
	//1.申请队列，如果队列不存在会自动创建，存在则跳过创建
	q, err := r.channel.QueueDeclare(
		r.QueueName,
		//是否持久化
		false,
		//是否自动删除
		false,
		//是否具有排他性
		false,
		//是否阻塞处理
		false,
		//额外的属性
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}

	//消费者流控，防止爆库
	r.channel.Qos(
		1,     //当前消费者一次能接受的最大消息数量
		0,     //服务器传递的最大容量（以八位字节为单位）
		false, //如果设置为true，对整个channel全局可用
	)

	//接收消息
	msgs, err := r.channel.Consume(
		q.Name, // queue
		//用来区分多个消费者
		"", // consumer
		//是否自动应答
		//这里要改掉，我们用手动应答
		false, // auto-ack
		//是否独有
		false, // exclusive
		//设置为true，表示 不能将同一个Conenction中生产者发送的消息传递给这个Connection中 的消费者
		false, // no-local
		//列是否阻塞
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		fmt.Println(err)
	}

	forever := make(chan bool)
	//启用协程处理消息
	go func() {
		for d := range msgs {
			//消息逻辑处理，可以自行设计逻辑
			log.Printf("Received a message: %s", d.Body)
			message := &datamodels.Message{}
			err := json.Unmarshal(d.Body, message)
			if err != nil {
				fmt.Println(err)
			}
			//插入订单
			_, err = orderService.InsertOrderByMessage(message)
			if err != nil {
				fmt.Println(err)
			}
			//扣除商品数量
			err = productService.SubNumberOne(message.ProductID)
			if err != nil {
				fmt.Println(err)
			}
			//如果为true表示确认所有未确认的消息，
			//为false表示确认当前消息
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
