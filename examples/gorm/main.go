package main

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败: ", err)
	}

	// AutoMigrate 会补齐表和字段，适合示例；正式项目仍应使用可追踪的迁移脚本。
	if err := db.AutoMigrate(&Product{}); err != nil {
		log.Fatal("迁移表结构失败: ", err)
	}

	// Create
	if err := db.Create(&Product{Code: "D42", Price: 100}).Error; err != nil {
		log.Fatal("创建商品失败: ", err)
	}

	// Read
	var product Product
	if err := db.First(&product, "code = ?", "D42").Error; err != nil {
		log.Fatal("查询商品失败: ", err)
	}

	// Update - 将 product 的 price 更新为 200
	db.Model(&product).Update("Price", 200)
	// Update - 更新多个字段
	db.Model(&product).Updates(Product{Price: 200, Code: "F42"}) // 仅更新非零值字段
	db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// Delete - 删除 product
	db.Delete(&product, 1)
}
