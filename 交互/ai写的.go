//要连接MongoDB,您需要使用`mongo-go-driver`包，说明如何使用sqlx连接到MongoDB:
import (
 "context"
 "log"

 "github.com/jmoiron/sqlx"
 "github.com/mattn/go-sqlite3"
 "go.mongodb.org/mongo-driver/mongo"
 "go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
 log.Println("Connecting to the database")
 db, err := sqlx.Connect("mongo", "mongodb://localhost:27017")
 if err != nil {
  log.Fatalln(err)
 }

 ctx := context.Background()
 client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
 if err != nil {
  log.Fatalln(err)
 }
 defer client.Disconnect(ctx)
 err = client.Connect(ctx)
 if err != nil {
  log.Fatalln(err)
 }

 collection := client.Database("test").Collection("users")
 users := []struct {
  ID       int    `json:"id"`
  Name     string `json:"name"`
  Age      int    `json:"age"`
 }{}
 // ...查询代码...
}

//或者使用mongo-driver连接

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// 设置MongoDB连接选项
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// 连接MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// 检查连接
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("成功连接到MongoDB！")

	// 断开连接
	err = client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("已断开与MongoDB的连接！")
}



