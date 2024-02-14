package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	// Importing a package with an underscore allows us to create the package
	// level variables and also execute the init function
	_ "github.com/mattn/go-sqlite3"
)

type Restaurant struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Stars   int      `json:"stars"`
	Address string   `json:"address"`
	State   string   `json:"state"`
	Hours   string   `json:"hours"`
	Chef    string   `json:"chef"`
	Staff   []string `json:"staff"`
	Photos  []string `json:"photos"`
	Website string   `json:"website"`
	Info    string   `json:"info"`
	Menus   []string `json:"menus"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", os.Getenv("DB"))
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
			chef TEXT NOT NULL,
			state TEXT NOT NULL,
			website TEXT NOT NULL,
			info TEXT NOT NULL
		);
	`)
	if err != nil {
		log.Fatal("Error creating table:", err)
	}

	// Load gin and HTML template support
	router := gin.Default()
	router.LoadHTMLGlob("templates/*.tmpl")

	// Route to get all restaurants
	router.GET("/api/v1/restaurants", GetRestaurantsHTML)

	// Route to get a single restaurant page by ID
	router.GET("/api/v1/restaurant/:id", GetRestaurantByIdHTML)

	// Route to create a new restaurant
	router.POST("/api/v1/restaurant/create", CreateRestaurantJSON)

	// Route to update a restaurant by ID
	router.PATCH("/api/v1/restaurant/update/:id", UpdateRestaurant)

	// Route to delete a restaurant by ID
	router.DELETE("/api/v1/restaurant/delete/:id", DeleteRestaurant)

	// Run the Gin server and check for errors
	if err := router.Run(":8083"); err != nil {
		log.Fatal("Error starting Gin server:", err)
	}
}

// GetRestaurantsHTML returns a list of all restaurants
func GetRestaurantsHTML(c *gin.Context) {
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
	c.HTML(http.StatusOK, "templates/restaurants.tmpl", gin.H{
		"title":       "Restaurants List",
		"restaurants": restaurants,
	})
}

// GetRestaurantByIdHTML returns info about a single restaurant
func GetRestaurantByIdHTML(c *gin.Context) {
	var restaurant Restaurant

	// Input parameters from URI
	id := c.Param("id")

	// Validate that the restaurant with this name exists in the database
	sqlStatement := `SELECT id, name, stars, address, state, website, chef, info
					 FROM restaurants
					 WHERE id = $1`

	row := db.QueryRow(sqlStatement, id)
	switch err := row.Scan(
		&restaurant.ID,
		&restaurant.Name,
		&restaurant.Stars,
		&restaurant.Address,
		&restaurant.State,
		&restaurant.Website,
		&restaurant.Chef,
		&restaurant.Info,
	); err {
	case sql.ErrNoRows:
		c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
	case nil:
		// Render HTML using the built-in HTML rendering
		c.HTML(http.StatusOK, "templates/restaurant.tmpl", gin.H{
			"ID":      restaurant.ID,
			"Name":    restaurant.Name,
			"Stars":   restaurant.Stars,
			"Address": restaurant.Address,
			"State":   restaurant.State,
			"Website": restaurant.Website,
			"Chef":    restaurant.Chef,
			"Info":    restaurant.Info,
		})
	default:
		panic(err)
	}
}

// CreateRestaurantJSON creates a new restaurant
func CreateRestaurantJSON(c *gin.Context) {
	name := c.PostForm("name")
	stars := c.PostForm("stars")
	address := c.PostForm("address")
	chef := c.PostForm("chef")

	result, readErr := db.Exec(
		`INSERT INTO restaurants (name, stars, address, chef)
	     VALUES (?, ?, ?, ?)`,
		name,
		stars,
		address,
		chef,
	)
	if readErr != nil {
		log.Println("Error inserting into database:", readErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	newID, err := result.LastInsertId()
	if err != nil {
		log.Fatalln("failed to insert into database")
	}
	c.JSON(http.StatusCreated, int(newID))
}

// UpdateRestaurant updates an existing restaurant by ID
func UpdateRestaurant(c *gin.Context) {
	id := c.PostForm("updateId")
	name := c.PostForm("updateName")
	stars := c.PostForm("updateStars")
	address := c.PostForm("updateAddress")
	chef := c.PostForm("updateChef")
	fmt.Println("fields:", name, stars, address, chef)

	if name == "" {
		log.Fatalln("name is blank")
	}
	if stars == "" {
		log.Fatalln("stars are blank")
	}
	if address == "" {
		log.Fatalln("address is blank")
	}
	if chef == "" {
		log.Fatalln("chef is blank")
	}

	var updatedRestaurant Restaurant

	// Check if the restaurant with the given ID exists
	var existingRestaurant Restaurant
	err := db.QueryRow("SELECT id, name, stars, address, chef, website, info FROM restaurants WHERE id = ?", id).
		Scan(
			&existingRestaurant.ID,
			&existingRestaurant.Name,
			&existingRestaurant.Stars,
			&existingRestaurant.Address,
			&existingRestaurant.Chef,
			&existingRestaurant.Website,
			&existingRestaurant.Info)
	if err != nil {
		log.Println("Error querying existing restaurant:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
		return
	}

	// Update the existing restaurant with the new data
	_, err = db.Exec(
		"UPDATE restaurants SET name = ?, stars = ?, address = ?, chef = ? WHERE id = ?",
		name,
		stars,
		address,
		chef,
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
	if id == "" {
		log.Fatalln("id provided to DeleteRestaurant() is nil")
	}

	statement, err := db.Prepare(`DELETE FROM restaurants WHERE id = ?`)
	if err != nil {
		log.Fatalln("failed to prepare delete statement", statement)
	}
	defer statement.Close()

	result, err := statement.Exec(id)
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

	deletedText := "Deleted"
	c.HTML(http.StatusOK, "templates/deleted.tmpl", gin.H{
		"deletedText": deletedText,
	})
}
