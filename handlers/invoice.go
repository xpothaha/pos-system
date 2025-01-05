package handlers

import (
	"fmt"
	"net/http"
	"pos-system/models"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/unrolled/render"
)

func CreateInvoiceHandler(db *gorm.DB, r *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			// Parse the form data
			if err := req.ParseForm(); err != nil {
				http.Error(w, "Error parsing form data", http.StatusBadRequest)
				return
			}

			// Extract invoice data
			customerName := req.FormValue("customer_name")
			saleDateStr := req.FormValue("sale_date")
			salesPerson := req.FormValue("sales_person")
			description := req.FormValue("description")
			note := req.FormValue("note")
			taxID := req.FormValue("tax_id")
			customerAddress := req.FormValue("customer_address")
			discountStr := req.FormValue("discount")

			// Validate and convert data types
			saleDate, err := time.Parse("2006-01-02T15:04", saleDateStr)
			if err != nil {
				http.Error(w, "Invalid sale date format", http.StatusBadRequest)
				fmt.Println("Error parsing date:", err)
				return
			}

			discount, err := strconv.ParseFloat(discountStr, 64)
			if err != nil {
				discount = 0 // Default to 0 if discount is not provided or invalid
			}

			// Generate Invoice Number
			var lastInvoice models.Invoice
			db.Order("id desc").First(&lastInvoice)
			nextInvoiceNumber := fmt.Sprintf("INV-%04d", lastInvoice.ID+1)

			// Create Invoice
			invoice := models.Invoice{
				InvoiceNumber:   nextInvoiceNumber,
				CustomerName:    customerName,
				SaleDate:        saleDate,
				SalesPerson:     salesPerson,
				Description:     description,
				Note:            note,
				TaxID:           taxID,
				CustomerAddress: customerAddress,
				Discount:        discount,
			}

			if err := db.Create(&invoice).Error; err != nil {
				http.Error(w, "Error creating invoice", http.StatusInternalServerError)
				return
			}

			// Get product data from form
			productIDs := req.Form["product_id[]"]
			itemNames := req.Form["item_name[]"]
			quantitiesStr := req.Form["quantity[]"]
			pricesStr := req.Form["price[]"]
			taxRatesStr := req.Form["tax_rate[]"]

			// Validate arrays have same length
			if len(productIDs) != len(itemNames) || len(productIDs) != len(quantitiesStr) ||
				len(productIDs) != len(pricesStr) || len(productIDs) != len(taxRatesStr) {
				http.Error(w, "Invalid product data: arrays length mismatch", http.StatusBadRequest)
				return
			}

			// Process each product
			for i := range productIDs {
				// Convert quantity to float64
				quantity, err := strconv.ParseFloat(quantitiesStr[i], 64)
				if err != nil {
					http.Error(w, fmt.Sprintf("Invalid quantity format for product %s", productIDs[i]), http.StatusBadRequest)
					return
				}

				// Convert price to float
				price, err := strconv.ParseFloat(pricesStr[i], 64)
				if err != nil {
					http.Error(w, fmt.Sprintf("Invalid price format for product %s", productIDs[i]), http.StatusBadRequest)
					return
				}

				// Convert tax rate to float
				taxRate, err := strconv.ParseFloat(taxRatesStr[i], 64)
				if err != nil {
					http.Error(w, fmt.Sprintf("Invalid tax rate format for product %s", productIDs[i]), http.StatusBadRequest)
					return
				}

				// Calculate tax and total
				tax := price * quantity * (taxRate / 100)
				total := price*quantity + tax

				// Create InvoiceItem
				invoiceItem := models.InvoiceItem{
					InvoiceID: invoice.ID,
					ProductID: productIDs[i],
					ItemName:  itemNames[i],
					Quantity:  quantity,
					Price:     price,
					TaxRate:   taxRate,
					Tax:       tax,
					Total:     total,
				}

				if err := db.Create(&invoiceItem).Error; err != nil {
					http.Error(w, "Error creating invoice item", http.StatusInternalServerError)
					return
				}
			}

			// Redirect to the invoice view page
			http.Redirect(w, req, "/invoice/"+strconv.Itoa(int(invoice.ID)), http.StatusSeeOther)
		} else if req.Method == http.MethodGet {
			r.HTML(w, http.StatusOK, "invoice", nil)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func ViewInvoiceHandler(db *gorm.DB, r *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		invoiceID, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Error(w, "Invalid invoice ID", http.StatusBadRequest)
			return
		}

		var invoice models.Invoice
		if err := db.Preload("Items").First(&invoice, invoiceID).Error; err != nil {
			http.Error(w, "Invoice not found", http.StatusNotFound)
			return
		}

		// Calculate totals
		var subtotal float64
		var totalTax float64
		for _, item := range invoice.Items {
			subtotal += item.Price * float64(item.Quantity)
			totalTax += item.Tax
		}

		// Apply discount
		total := subtotal + totalTax - invoice.Discount

		// Create a data structure to pass to the template
		data := struct {
			Invoice  models.Invoice
			Subtotal float64
			TotalTax float64
			Total    float64
		}{
			Invoice:  invoice,
			Subtotal: subtotal,
			TotalTax: totalTax,
			Total:    total,
		}

		r.HTML(w, http.StatusOK, "view_invoice", data)
	}
}
