package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Product struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Price    int    `json:"price"`
}

var db *pgxpool.Pool

func main() {
	ctx := context.Background()
	conn, err := pgxpool.New(ctx, "postgres://user:password@localhost:5432/mydb")
	if err != nil {
		log.Fatal("DB connection error:", err)
	}
	db = conn
	defer db.Close()

	router := gin.Default()
	router.GET("/products", getProductsHandler)

	log.Println("🚀 Server running on http://localhost:8080")
	router.Run(":8080")
}

func getProductsHandler(c *gin.Context) {
	start := time.Now()
	category := c.Query("category")
	minPriceStr := c.Query("min_price")
	maxPriceStr := c.Query("max_price")
	sort := c.Query("sort")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	query := `
		SELECT p.id, p.name, c.name AS category, p.price
		FROM products p
		JOIN categories c ON p.category_id = c.id
		WHERE 1=1
	`
	args := []interface{}{}
	argIndex := 1

	if category != "" {
		query += " AND c.name = $" + strconv.Itoa(argIndex)
		args = append(args, category)
		argIndex++
	}
	if minPriceStr != "" {
		query += " AND p.price >= $" + strconv.Itoa(argIndex)
		if val, err := strconv.Atoi(minPriceStr); err == nil {
			args = append(args, val)
			argIndex++
		}
	}
	if maxPriceStr != "" {
		query += " AND p.price <= $" + strconv.Itoa(argIndex)
		if val, err := strconv.Atoi(maxPriceStr); err == nil {
			args = append(args, val)
			argIndex++
		}
	}

	if sort == "price_asc" {
		query += " ORDER BY p.price ASC"
	} else if sort == "price_desc" {
		query += " ORDER BY p.price DESC"
	}

	query += " LIMIT $" + strconv.Itoa(argIndex)
	limit, _ := strconv.Atoi(limitStr)
	args = append(args, limit)
	argIndex++

	query += " OFFSET $" + strconv.Itoa(argIndex)
	offset, _ := strconv.Atoi(offsetStr)
	args = append(args, offset)

	rows, err := db.Query(context.Background(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.Price)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		products = append(products, p)
	}

	elapsed := time.Since(start)
	c.Header("X-Query-Time", elapsed.String())

	c.JSON(http.StatusOK, products)
}
