package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

type User struct {
	ID        uint   `gorm:"primaryKey"`
	FullName  string `json:"full_name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required"`
	Phone     string `json:"phone" validate:"required"`
	Role      string `json:"role"`
	CreatedAt time.Time
}

type Car struct {
	ID           uint    `gorm:"primaryKey"`
	Brand        string  `json:"brand" validate:"required"`
	Model        string  `json:"model" validate:"required"`
	Year         int     `json:"year" validate:"required"`
	LicensePlate string  `json:"license_plate" validate:"required"`
	PricePerDay  float64 `json:"price_per_day" validate:"required"`
	Status       string  `json:"status"`
	CreatedAt    time.Time
}

type Rental struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `json:"user_id" validate:"required"`
	CarID      uint      `json:"car_id" validate:"required"`
	StartDate  time.Time `json:"start_date" validate:"required"`
	EndDate    time.Time `json:"end_date" validate:"required"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"`
	CreatedAt  time.Time
}

type Payment struct {
	ID            uint    `gorm:"primaryKey"`
	RentalID      uint    `json:"rental_id" validate:"required"`
	Amount        float64 `json:"amount" validate:"required"`
	PaymentMethod string  `json:"payment_method" validate:"required"`
	Status        string  `json:"status"`
	CreatedAt     time.Time
}

func connectDatabase() {
	var err error

	// Load .env file
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get database connection details from the .env file
	dsn := os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") + "@tcp(" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + ")/" + os.Getenv("DB_NAME") + "?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("Database connected")

	// Migrate the models
	db.AutoMigrate(&User{}, &Car{}, &Rental{}, &Payment{})

	// Create a default admin if not exists
	adminEmail := os.Getenv("ADMIN_EMAIL")
	var admin User
	db.Where("email = ?", adminEmail).First(&admin)
	if admin.ID == 0 {
		hashedPassword := os.Getenv("ADMIN_PASSWORD")
		db.Create(&User{
			FullName: "Admin Default",
			Email:    adminEmail,
			Password: hashedPassword,
			Phone:    "1234567890",
			Role:     "admin",
		})
		log.Println("Admin default created.")
	}
}

func main() {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3001, http://localhost:3002, https://frontend-chi-seven-62.vercel.app", 
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
		ExposeHeaders:    "Authorization",
	}))

	connectDatabase()

	app.Post("/register", registerUser)
	app.Post("/login", loginUser)

	// Protected routes middleware JWT
	app.Use(jwtware.New(jwtware.Config{
		SigningKey:  []byte(os.Getenv("JWT_SECRET_KEY")),
		TokenLookup: "header:Authorization",
		AuthScheme:  "Bearer",
	}))

	// CRUD Routes
	app.Get("/users", getUsers)
	app.Get("/users/:id", getUser)
	app.Post("/users", createUser) // Menambahkan rute POST /users untuk membuat user baru
	app.Put("/users/:id", updateUser)
	app.Delete("/users/:id", deleteUser)
	// Endpoint untuk mendapatkan data user yang sedang login
	app.Get("/users/me", getMe)

	app.Get("/cars", getCars)
	app.Get("/cars/:id", getCar)
	app.Post("/cars", addCar)
	app.Put("/cars/:id", updateCar)
	app.Delete("/cars/:id", deleteCar)

	app.Get("/rentals", getRentals)
	app.Get("/rentals/:id", getRental)
	app.Post("/rentals", createRental)
	app.Put("/rentals/:id", updateRental)
	app.Delete("/rentals/:id", deleteRental)

	app.Get("/payments", getPayments)
	app.Get("/payments/:id", getPayment)
	app.Post("/payments", createPayment)
	app.Put("/payments/:id", updatePayment)
	app.Delete("/payments/:id", deletePayment)

	log.Fatal(app.Listen(":3000"))
}

// Handlers untuk setiap endpoint tetap sama dengan validasi tambahan

// User Handlers
func registerUser(c *fiber.Ctx) error {
	user := new(User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Set role default sebagai 'customer'
	user.Role = "customer"
	db.Create(&user)
	return c.JSON(user)
}

func loginUser(c *fiber.Ctx) error {
	input := new(User)
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	var user User
	db.Where("email = ? AND password = ?", input.Email, input.Password).First(&user)
	if user.ID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create token"})
	}

	return c.JSON(fiber.Map{"token": t, "role": user.Role})
}

// CRUD Handlers lainnya tetap sama seperti sebelumnya
// (getUsers, getCars, addCar, dan sebagainya)

func getUsers(c *fiber.Ctx) error {
	var users []User
	db.Find(&users)
	return c.JSON(users)
}

func getUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var user User
	result := db.First(&user, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	return c.JSON(user)
}

func createUser(c *fiber.Ctx) error {
	user := new(User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Set role default sebagai 'customer'
	user.Role = "customer"
	db.Create(&user) // Simpan user ke dalam database

	return c.JSON(user) // Kembalikan data user yang baru dibuat sebagai respons
}

func updateUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var user User
	result := db.First(&user, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	db.Save(&user)
	return c.JSON(user)
}

func deleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	var user User
	result := db.First(&user, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	db.Delete(&user)
	return c.SendStatus(fiber.StatusNoContent)
}

func getMe(c *fiber.Ctx) error {
	// Mengambil user_id dari token JWT yang sudah ter-parse oleh middleware
	userID := c.Locals("user_id").(float64) // Casting ke float64, sesuai tipe yang ada di token JWT

	var user User
	result := db.First(&user, uint(userID)) // Pastikan casting ke uint yang benar
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Mengembalikan data user yang sedang login
	return c.JSON(user)
}

// Cars Handlers
func getCars(c *fiber.Ctx) error {
	var cars []Car
	db.Find(&cars)
	return c.JSON(cars)
}

func getCar(c *fiber.Ctx) error {
	id := c.Params("id")
	var car Car
	result := db.First(&car, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Car not found"})
	}
	return c.JSON(car)
}

func addCar(c *fiber.Ctx) error {
	car := new(Car)
	if err := c.BodyParser(car); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	db.Create(&car)
	return c.JSON(car)
}

func updateCar(c *fiber.Ctx) error {
	id := c.Params("id")
	var car Car
	result := db.First(&car, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Car not found"})
	}
	if err := c.BodyParser(&car); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	db.Save(&car)
	return c.JSON(car)
}

func deleteCar(c *fiber.Ctx) error {
	id := c.Params("id")
	var car Car
	result := db.First(&car, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Car not found"})
	}
	db.Delete(&car)
	return c.SendStatus(fiber.StatusNoContent)
}

// Rentals Handlers
func getRentals(c *fiber.Ctx) error {
	var rentals []Rental
	db.Find(&rentals)
	return c.JSON(rentals)
}

func getRental(c *fiber.Ctx) error {
	id := c.Params("id")
	var rental Rental
	result := db.First(&rental, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Rental not found"})
	}
	return c.JSON(rental)
}

func createRental(c *fiber.Ctx) error {
	rental := new(Rental)
	if err := c.BodyParser(rental); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Ensure that user with the provided ID exists
	var user User
	if result := db.First(&user, rental.UserID); result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Get the car related to the rental
	carID := rental.CarID
	var car Car
	db.First(&car, carID)
	if car.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Car not found"})
	}

	// Calculate rental duration in days
	duration := rental.EndDate.Sub(rental.StartDate).Hours() / 24

	// Calculate total rental price
	rental.TotalPrice = car.PricePerDay * duration

	// Set rental status to "unpaid"
	rental.Status = "unpaid"

	// Save rental to the database
	if result := db.Create(&rental); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create rental"})
	}

	// Check if rental ID is set correctly
	if rental.ID == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create rental, ID is missing"})
	}

	// Create payment entry with status "unpaid"
	payment := Payment{
		RentalID:      rental.ID,
		Amount:        rental.TotalPrice,
		PaymentMethod: "", // Can be filled when payment is made
		Status:        "unpaid",
	}
	if result := db.Create(&payment); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create payment"})
	}

	return c.JSON(rental) // Send rental back with ID
}

func updateRental(c *fiber.Ctx) error {
	id := c.Params("id")
	var rental Rental
	result := db.First(&rental, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Rental not found"})
	}
	if err := c.BodyParser(&rental); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	db.Save(&rental)
	return c.JSON(rental)
}

func deleteRental(c *fiber.Ctx) error {
	id := c.Params("id")
	var rental Rental
	result := db.First(&rental, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Rental not found"})
	}
	db.Delete(&rental)
	return c.SendStatus(fiber.StatusNoContent)
}

// Payments Handlers
func getPayments(c *fiber.Ctx) error {
	var payments []Payment
	db.Find(&payments)
	return c.JSON(payments)
}

func getPayment(c *fiber.Ctx) error {
	id := c.Params("id")
	var payment Payment
	result := db.First(&payment, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Payment not found"})
	}
	return c.JSON(payment)
}

func createPayment(c *fiber.Ctx) error {
	payment := new(Payment)
	if err := c.BodyParser(payment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	// Get the rental related to the payment
	var rental Rental
	db.First(&rental, payment.RentalID)
	if rental.ID == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Rental not found"})
	}

	// Check if the rental is already paid
	if rental.Status == "paid" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Rental already paid"})
	}

	// Change rental and payment status to "paid"
	rental.Status = "paid"
	payment.Status = "paid"

	// Save the updated status for rental and payment
	db.Save(&rental)
	db.Create(&payment) // Record the payment in the database

	return c.JSON(payment)
}

func updatePayment(c *fiber.Ctx) error {
	id := c.Params("id")
	var payment Payment
	result := db.First(&payment, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Payment not found"})
	}
	if err := c.BodyParser(&payment); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	db.Save(&payment)
	return c.JSON(payment)
}

func deletePayment(c *fiber.Ctx) error {
	id := c.Params("id")
	var payment Payment
	result := db.First(&payment, id)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Payment not found"})
	}
	db.Delete(&payment)
	return c.SendStatus(fiber.StatusNoContent)
}
