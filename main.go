package main

import (
	"database/sql"
	"log"
	"net/http"

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
	router.GET("/api/v1/restaurants/get", GetRestaurants)

	// Route to get a single restaurant page by ID in the database
	router.GET("/api/v1/restaurant/:id", GetRestaurantByID)

	// Route to create a new restaurant
	router.POST("/api/v1/restaurants/create", CreateRestaurant)

	// Route to update a restaurant by ID
	router.PUT("/api/v1/restaurants/update/:id", UpdateRestaurant)

	// Route to delete a restaurant by ID
	router.DELETE("/api/v1/restaurants/delete/:id", DeleteRestaurant)

	// Run the Gin server and check for errors
	if err := router.Run(":8083"); err != nil {
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
	c.HTML(http.StatusOK, "templates/restaurants.tmpl", gin.H{
		"title":       "Restaurants List",
		"restaurants": restaurants,
	})
}

// GetRestaurantByID returns info about a single restaurant
func GetRestaurantByID(c *gin.Context) {
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

// CreateRestaurant creates a new restaurant
func CreateRestaurant(c *gin.Context) {
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
