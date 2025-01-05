package main

import (
        "fmt"
        "html/template"
        "log"
        "net/http"
        "os"
        "pos-system/handlers"
        "pos-system/models"

        "github.com/gorilla/mux"
        "github.com/jinzhu/gorm"
        _ "github.com/jinzhu/gorm/dialects/postgres"
        "github.com/unrolled/render"
)

func main() {
        // ใช้ค่าที่กำหนดไว้โดยตรง แทนการใช้ os.Getenv ในการทดสอบ
        // เมื่อใช้งานจริง ควรกลับไปใช้ os.Getenv เพื่อความปลอดภัย
        dbUser := "postgres"
        dbPass := "147258369"
        dbName := "pos_db"
        dbHost := "localhost"
        dbPort := "5432"

        // ตรวจสอบว่า environment variable ถูกตั้งค่าหรือไม่ (สำหรับ Production)
        if os.Getenv("DB_USER") != "" {
                dbUser = os.Getenv("DB_USER")
        }
        if os.Getenv("DB_PASSWORD") != "" {
                dbPass = os.Getenv("DB_PASSWORD")
        }
        if os.Getenv("DB_NAME") != "" {
                dbName = os.Getenv("DB_NAME")
        }
        if os.Getenv("DB_HOST") != "" {
                dbHost = os.Getenv("DB_HOST")
        }
        if os.Getenv("DB_PORT") != "" {
                dbPort = os.Getenv("DB_PORT")
        }

        connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)
        fmt.Println("Connection String:", connStr) // แสดง connection string สำหรับ debug

        db, err := gorm.Open("postgres", connStr)
        if err != nil {
                log.Fatalf("failed to connect database: %v", err) // แสดง error อย่างละเอียดและ exit
        }
        defer db.Close()

        // ตรวจสอบการเชื่อมต่อโดยการ Ping ฐานข้อมูล
        if err := db.DB().Ping(); err != nil {
                log.Fatalf("failed to ping database: %v", err)
        } else {
                fmt.Println("Successfully connected to database!")
        }

        db.AutoMigrate(&models.Invoice{}, &models.InvoiceItem{}, &models.Product{})

        // Define custom functions for templates
        funcMap := template.FuncMap{
                "add": func(a, b float64) float64 {
                        return a + b
                },
                "mul": func(a, b float64) float64 {
                        return a * b
                },
        }

        r := render.New(render.Options{
                Directory:     "templates",
                Extensions:    []string{".html"},
                IndentJSON:    true,
                Funcs:         []template.FuncMap{funcMap}, // Register custom functions
        })

        router := mux.NewRouter()

        router.HandleFunc("/invoice", handlers.CreateInvoiceHandler(db, r)).Methods("GET", "POST")
        router.HandleFunc("/invoice/{id}", handlers.ViewInvoiceHandler(db, r)).Methods("GET")

        fmt.Println("Starting server on :8080")
        log.Fatal(http.ListenAndServe(":8080", router))
}