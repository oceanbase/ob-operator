package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/oceanbase/ob-operator/todo/internal"
)

func main() {
	dsn, err := internal.NewDSN()
	if err != nil {
		panic("Failed to create a DSN: " + err.Error())
	}
	log.Println("DSN: ", dsn.String())
	db, err := gorm.Open(mysql.Open(dsn.String()), &gorm.Config{})
	if err != nil {
		panic("Failed to open a database connection: " + err.Error())
	}
	err = db.AutoMigrate(&internal.Todo{})
	if err != nil {
		panic("Failed to auto migrate the Todo model: " + err.Error())
	}
	// Add some initial data if the table is empty
	var count int64
	db.Unscoped().Model(&internal.Todo{}).Count(&count)
	if count == 0 {
		res := db.CreateInBatches(internal.InitialTodos, 5)
		if res.Error != nil {
			panic("Failed to create initial todos: " + res.Error.Error())
		}
	}

	r := gin.Default()
	r.Use(cors.Default())
	r.Use(static.Serve("/", static.LocalFile("ui/dist", false)))

	r.GET("/api/todos", func(c *gin.Context) {
		todos := []internal.Todo{}
		tx := db.Find(&todos)
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": tx.Error.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"todos": todos,
		})
	})

	r.PUT("/api/todos", func(c *gin.Context) {
		var todo internal.Todo
		err := c.BindJSON(&todo)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		tx := db.Create(&todo)
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": tx.Error.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"todo": todo,
		})
	})

	r.DELETE("/api/todos/:id", func(c *gin.Context) {
		id := c.Param("id")
		tx := db.Delete(&internal.Todo{}, id)
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": tx.Error.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{})
	})

	r.PATCH("/api/todos/:id", func(c *gin.Context) {
		id := c.Param("id")
		var todo internal.EditTodo
		err := c.BindJSON(&todo)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		t := map[string]interface{}{}
		if todo.Title != "" {
			t["title"] = todo.Title
		}
		if todo.Description != "" {
			t["description"] = todo.Description
		}
		if todo.FinishedAt != nil {
			t["finished_at"] = todo.FinishedAt
		} else if todo.ClearFinishedAt {
			t["finished_at"] = nil
		}
		tx := db.Model(&internal.Todo{}).Where("id = ?", id).Updates(t)
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": tx.Error.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"todo": todo,
		})
	})

	r.Run(":20031")
}
