package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/mattn/go-sqlite3"
)

type Restaurant struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Stars   int    `json:"stars"`
	Address string `json:"address"`
	Chef    string `json:"chef"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./nocodb/restaurants.db")
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	// Create restaurants table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS restaurants (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			stars INTEGER NOT NULL,
			address TEXT NOT NULL,
			chef TEXT NOT NULL
		);
	`)
	if err != nil {
		log.Fatal("Error creating table:", err)
	}

	// Load gin and HTML template support
	r := gin.Default()
	r.LoadHTMLGlob("*.html")

	// Route to get all restaurants
	r.GET("/api/v1/restaurants/get", GetRestaurants)

	// Route to create a new restaurant
	r.POST("/api/v1/restaurants/create", CreateRestaurant)

	// Route to update a restaurant by ID
	r.PUT("/api/v1/restaurants/update/:id", UpdateRestaurant)

	// Route to delete a restaurant by ID
	r.DELETE("/api/v1/restaurants/delete/:id", DeleteRestaurant)

	// Run the Gin server and check for errors
	if err := r.Run(":8081"); err != nil {
		log.Fatal("Error starting Gin server:", err)
	}
}

// GetRestaurants returns a list of all restaurants
func GetRestaurants(c *gin.Context) {
	rows, err := db.Query("SELECT id, name, stars, address, chef FROM restaurants")
	if err != nil {
		log.Println("Error retrieving restaurants:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	defer rows.Close()

	var restaurants []Restaurant
	for rows.Next() {
		var restaurant Restaurant
		err := rows.Scan(
			&restaurant.ID,
			&restaurant.Name,
			&restaurant.Stars,
			&restaurant.Address,
			&restaurant.Chef,
		)
		if err != nil {
			log.Println("Error scanning row:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		restaurants = append(restaurants, restaurant)
	}

	// Render HTML using the built-in HTML rendering
	c.HTML(http.StatusOK, "restaurants.html", gin.H{
		"title":       "Restaurants List",
		"restaurants": restaurants,
	})
	// c.JSON(http.StatusOK, restaurants)
}

// CreateRestaurant creates a new restaurant
func CreateRestaurant(c *gin.Context) {
	var newRestaurant Restaurant

	// Read the request body
	body, readErr := io.ReadAll(c.Request.Body)
	log.Println("request body contains: ", string(body))
	// log.Println("name: ", newRestaurant.Name)
	// log.Println("address: ", newRestaurant.Address)
	// log.Println("chef: ", newRestaurant.Chef)
	// log.Println("stars: ", newRestaurant.Stars)
	// log.Println("id: ", newRestaurant.ID)
	if readErr != nil {
		log.Println("Error reading request body:", readErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// Unmarshal the JSON data into the newRestaurant struct
	if unMarshallErr := json.Unmarshal(body, &newRestaurant); readErr != nil {
		log.Println("Error unmarshalling JSON:", unMarshallErr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	} else {
		log.Println("unmarshalled json ok")
		log.Println("c.ShouldBindJSON(&newRestaurant)", c.ShouldBindJSON(&newRestaurant))
		log.Printf("Unmarshalled Data: %+v", newRestaurant)
	}

	// Validate that we can parse the JSON before writing to database
	if bindErr := c.ShouldBindJSON(&newRestaurant); bindErr != nil {
		log.Println("Error binding JSON:", bindErr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	result, readErr := db.Exec(
		"INSERT INTO restaurants (name, stars, address, chef) VALUES (?, ?, ?, ?)",
		newRestaurant.Name,
		newRestaurant.Stars,
		newRestaurant.Address,
		newRestaurant.Chef,
	)
	if readErr != nil {
		log.Println("Error inserting into database:", readErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	newID, _ := result.LastInsertId()
	newRestaurant.ID = int(newID)

	c.JSON(http.StatusCreated, newRestaurant)
}

// UpdateRestaurant updates an existing restaurant by ID
func UpdateRestaurant(c *gin.Context) {
	id := c.Param("id")

	var updatedRestaurant Restaurant
	if err := c.ShouldBindJSON(&updatedRestaurant); err != nil {
		log.Println("Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Check if the restaurant with the given ID exists
	var existingRestaurant Restaurant
	err := db.QueryRow("SELECT id, name, stars, address, chef FROM restaurants WHERE id = ?", id).
		Scan(&existingRestaurant.ID, &existingRestaurant.Name, &existingRestaurant.Stars, &existingRestaurant.Address, &existingRestaurant.Chef)
	if err != nil {
		log.Println("Error querying existing restaurant:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
		return
	}

	// Update the existing restaurant with the new data
	_, err = db.Exec(
		"UPDATE restaurants SET name = ?, stars = ?, address = ?, chef = ? WHERE id = ?",
		updatedRestaurant.Name,
		updatedRestaurant.Stars,
		updatedRestaurant.Address,
		updatedRestaurant.Chef,
		id,
	)
	if err != nil {
		log.Println("Error updating restaurant:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// Return the updated restaurant
	updatedRestaurant.ID = existingRestaurant.ID
	c.JSON(http.StatusOK, updatedRestaurant)
}

// DeleteRestaurant deletes a restaurant by ID
func DeleteRestaurant(c *gin.Context) {
	id := c.Param("id")
	log.Println("deleting id:", id)

	result, err := db.Exec("DELETE FROM restaurants WHERE id = ?", id)
	if err != nil {
		log.Println("Error deleting restaurant:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Restaurant deleted successfully"})
}
