package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Task struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	Status      string     `json:"status"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type App struct {
	DB       *sql.DB
	Sessions *session.Store
}

const sessionUserKey = "user_id"

func main() {
	if err := godotenv.Overload(".env"); err != nil {
		log.Println(".env yüklenemedi, ortam değişkenleri doğrudan kullanılacak:", err)
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL ortam değişkeni tanımlı olmalı (örn. Supabase connection string)")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	if err := migrate(db); err != nil {
		log.Fatal(err)
	}

	isProduction := os.Getenv("APP_ENV") == "production"

	store := session.New(session.Config{
		Expiration:     24 * time.Hour,
		CookieSecure:   isProduction,
		CookieHTTPOnly: true,
		CookieSameSite: "None",
	})

	app := &App{
		DB:       db,
		Sessions: store,
	}

	f := fiber.New()

	f.Use(logger.New())
	f.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173,http://localhost:5174,https://task-app-ett8.vercel.app",
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	}))

	api := f.Group("/api")

	api.Post("/auth/register", app.handleRegister)
	api.Post("/auth/login", app.handleLogin)
	api.Post("/auth/logout", app.handleLogout)
	api.Post("/auth/reset-password", app.handleResetPassword)

	api.Get("/me", app.withAuth(app.handleGetMe))
	api.Put("/me", app.withAuth(app.handleUpdateMe))

	api.Get("/tasks", app.withAuth(app.handleListTasks))
	api.Post("/tasks", app.withAuth(app.handleCreateTask))
	api.Put("/tasks/:id", app.withAuth(app.handleUpdateTask))
	api.Patch("/tasks/:id/status", app.withAuth(app.handleUpdateTaskStatus))
	api.Delete("/tasks/:id", app.withAuth(app.handleDeleteTask))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server %s portunda dinleniyor...", port)
	if err := f.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}

func migrate(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tasks (
	id SERIAL PRIMARY KEY,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	title TEXT NOT NULL,
	body TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'pending',
	completed_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`
	_, err := db.Exec(schema)
	return err
}

func (a *App) getSession(c *fiber.Ctx) (*session.Session, error) {
	return a.Sessions.Get(c)
}

func (a *App) withAuth(next fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sess, err := a.getSession(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "session error"})
		}
		userID := sess.Get(sessionUserKey)
		if userID == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		c.Locals(sessionUserKey, userID.(int64))
		return next(c)
	}
}

func (a *App) handleRegister(c *fiber.Ctx) error {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "geçersiz istek"})
	}
	if body.Name == "" || body.Email == "" || body.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name, email, password zorunlu"})
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "şifre hatası"})
	}
	var id int64
	err = a.DB.QueryRow(
		`INSERT INTO users (name, email, password_hash)
         VALUES ($1, $2, $3)
         ON CONFLICT (email)
         DO UPDATE SET
           name = EXCLUDED.name,
           password_hash = EXCLUDED.password_hash,
           updated_at = NOW()
         RETURNING id`,
		body.Name, body.Email, string(hash),
	).Scan(&id)
	if err != nil {
		log.Printf("kayıt/güncelleme hatası: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "kayıt/güncelleme hatası"})
	}

	sess, err := a.getSession(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "session error"})
	}
	sess.Set(sessionUserKey, id)
	if err := sess.Save(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "session kaydedilemedi"})
	}

	return c.JSON(fiber.Map{"id": id, "name": body.Name, "email": body.Email})
}

func (a *App) handleLogin(c *fiber.Ctx) error {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "geçersiz istek"})
	}
	if body.Email == "" || body.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email ve password zorunlu"})
	}

	var u User
	err := a.DB.QueryRow(
		"SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE email = $1",
		body.Email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "geçersiz email veya şifre"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(body.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "geçersiz email veya şifre"})
	}

	sess, err := a.getSession(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "session error"})
	}
	sess.Set(sessionUserKey, u.ID)
	if err := sess.Save(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "session kaydedilemedi"})
	}

	return c.JSON(fiber.Map{
		"id":    u.ID,
		"name":  u.Name,
		"email": u.Email,
	})
}

func (a *App) handleLogout(c *fiber.Ctx) error {
	sess, err := a.getSession(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "session error"})
	}
	if err := sess.Destroy(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "logout başarısız"})
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (a *App) handleResetPassword(c *fiber.Ctx) error {
	var body struct {
		Email       string `json:"email"`
		NewPassword string `json:"new_password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "geçersiz istek"})
	}
	if body.Email == "" || body.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email ve new_password zorunlu"})
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "şifre hatası"})
	}
	res, err := a.DB.Exec(
		"UPDATE users SET password_hash = $1, updated_at = NOW() WHERE email = $2",
		string(hash), body.Email,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "güncelleme hatası"})
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "bu email ile kayıtlı kullanıcı yok"})
	}
	return c.JSON(fiber.Map{"ok": true, "message": "şifre güncellendi, giriş yapabilirsin"})
}

func (a *App) handleGetMe(c *fiber.Ctx) error {
	userID := c.Locals(sessionUserKey).(int64)
	var u User
	err := a.DB.QueryRow(
		"SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE id = $1",
		userID,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "kullanıcı bulunamadı"})
	}
	return c.JSON(fiber.Map{
		"id":    u.ID,
		"name":  u.Name,
		"email": u.Email,
	})
}

func (a *App) handleUpdateMe(c *fiber.Ctx) error {
	userID := c.Locals(sessionUserKey).(int64)
	var body struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "geçersiz istek"})
	}
	if body.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name zorunlu"})
	}
	_, err := a.DB.Exec(
		"UPDATE users SET name = $1, updated_at = NOW() WHERE id = $2",
		body.Name, userID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "güncelleme hatası"})
	}
	return c.JSON(fiber.Map{"ok": true, "name": body.Name})
}

func (a *App) handleListTasks(c *fiber.Ctx) error {
	userID := c.Locals(sessionUserKey).(int64)
	rows, err := a.DB.Query(
		"SELECT id, title, body, status, completed_at, created_at, updated_at FROM tasks WHERE user_id = $1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		log.Printf("task listeleme hatası: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "listeleme hatası (detay için backend loguna bak)"})
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Body, &t.Status, &t.CompletedAt, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "okuma hatası"})
		}
		tasks = append(tasks, t)
	}
	return c.JSON(tasks)
}

func (a *App) handleCreateTask(c *fiber.Ctx) error {
	userID := c.Locals(sessionUserKey).(int64)
	var body struct {
		Title  string `json:"title"`
		Body   string `json:"body"`
		Status string `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "geçersiz istek"})
	}
	if body.Title == "" || body.Body == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title ve body zorunlu"})
	}
	if body.Status == "" {
		body.Status = "pending"
	}

	var id int64
	err := a.DB.QueryRow(
		"INSERT INTO tasks (user_id, title, body, status) VALUES ($1, $2, $3, $4) RETURNING id",
		userID, body.Title, body.Body, body.Status,
	).Scan(&id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "task oluşturulamadı"})
	}
	return c.JSON(fiber.Map{
		"id":     id,
		"title":  body.Title,
		"body":   body.Body,
		"status": body.Status,
	})
}

func (a *App) handleUpdateTask(c *fiber.Ctx) error {
	userID := c.Locals(sessionUserKey).(int64)
	id := c.Params("id")
	var body struct {
		Title  string `json:"title"`
		Body   string `json:"body"`
		Status string `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "geçersiz istek"})
	}
	if body.Title == "" || body.Body == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title ve body zorunlu"})
	}
	if body.Status == "" {
		body.Status = "pending"
	}

	_, err := a.DB.Exec(
		"UPDATE tasks SET title = $1, body = $2, status = $3, updated_at = NOW() WHERE id = $4 AND user_id = $5",
		body.Title, body.Body, body.Status, id, userID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "task güncellenemedi"})
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (a *App) handleUpdateTaskStatus(c *fiber.Ctx) error {
	userID := c.Locals(sessionUserKey).(int64)
	id := c.Params("id")
	var body struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "geçersiz istek"})
	}
	if body.Status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status zorunlu"})
	}

	var query string
	if body.Status == "done" || body.Status == "completed" {
		query = "UPDATE tasks SET status = $1, completed_at = NOW(), updated_at = NOW() WHERE id = $2 AND user_id = $3"
	} else {
		query = "UPDATE tasks SET status = $1, completed_at = NULL, updated_at = NOW() WHERE id = $2 AND user_id = $3"
	}

	_, err := a.DB.Exec(query, body.Status, id, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "durum güncellenemedi"})
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (a *App) handleDeleteTask(c *fiber.Ctx) error {
	userID := c.Locals(sessionUserKey).(int64)
	id := c.Params("id")
	_, err := a.DB.Exec(
		"DELETE FROM tasks WHERE id = $1 AND user_id = $2",
		id, userID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "task silinemedi"})
	}
	return c.JSON(fiber.Map{"ok": true})
}
