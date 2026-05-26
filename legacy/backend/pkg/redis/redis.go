package redis

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

func FlushDB(port string) {
	// 创建一个 Redis 客户端
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:" + port, // 默认连接本地 Redis
		Password: "",                  // 如果没设置密码，可以留空
		DB:       0,                   // 使用默认的 DB 0
	})

	// 检查是否已连接到 Redis
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("无法连接到 Redis: %v", err)
	}

	fmt.Println(rdb.DBSize(ctx))
	// 执行 FLUSHDB 命令
	err = rdb.FlushDB(ctx).Err()
	if err != nil {
		log.Fatalf("执行 FLUSHDB 失败: %v", err)
	}

	fmt.Println(rdb.DBSize(ctx))

}
