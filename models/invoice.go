package models

import "time"

type Invoice struct {
        ID              uint      `gorm:"primary_key"`
        InvoiceNumber   string    `gorm:"unique_index;not null"`
        CustomerName    string    `gorm:"not null"`
        TaxID           string    // New field
        CustomerAddress string    // New field
        SaleDate        time.Time `gorm:"not null"`
        SalesPerson     string    `gorm:"not null"`
        Description     string
        Note            string
        Discount        float64   // New field
        Items           []InvoiceItem `gorm:"foreignkey:InvoiceID"`
        CreatedAt       time.Time
        UpdatedAt       time.Time
}

type InvoiceItem struct {
        ID        uint    `gorm:"primary_key"`
        InvoiceID uint    `gorm:"not null"`
        ProductID string    `gorm:"not null"`
        ItemName  string    `gorm:"not null"`
        Quantity  int     `gorm:"not null"`
        Price     float64 `gorm:"not null"`
        TaxRate   float64 // New field
        Tax       float64 // New field
        Total     float64 // เพิ่ม field นี้เข้าไป
}

type Product struct {
        ID        string  `gorm:"primary_key"`
        Name      string
        Price     float64
        Stock     int
        CreatedAt time.Time
        UpdatedAt time.Time
}