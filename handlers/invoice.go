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
            saleDateStr := req.FormValue("sale_date") // ประกาศที่นี่
            salesPerson := req.FormValue("sales_person")
            description := req.FormValue("description")
            note := req.FormValue("note")
            taxID := req.FormValue("tax_id")
            customerAddress := req.FormValue("customer_address")
            discountStr := req.FormValue("discount")

            productsData := req.FormValue("products_data")
            var products []map[string]interface{}
            if err := json.Unmarshal([]byte(productsData), &products); err != nil {
                http.Error(w, "Invalid products data", http.StatusBadRequest)
                return
            }
        
            for _, productData := range products {
                // Handle quantity as string first
                quantityStr, ok := productData["quantity"].(string)
                if !ok {
                    http.Error(w, "Invalid quantity format: not a string", http.StatusBadRequest)
                    return
                }
        
                quantity, err := strconv.Atoi(quantityStr)
                if err != nil {
                    http.Error(w, "Invalid quantity format: not a valid integer", http.StatusBadRequest)
                    return
                }
                //Get productID, itemName, price, taxRate, tax, total from productData
                        productId, ok := productData["productId"].(string)
                if !ok {
                    http.Error(w, "Invalid product ID format", http.StatusBadRequest)
                    return
                }
                itemName, ok := productData["itemName"].(string)
                if !ok {
                    http.Error(w, "Invalid item name format", http.StatusBadRequest)
                    return
                }
                priceStr, ok := productData["price"].(string)
                if !ok {
                    http.Error(w, "Invalid price format: not a string", http.StatusBadRequest)
                    return
                }
        
                price, err := strconv.ParseFloat(priceStr, 64)
                if err != nil {
                    http.Error(w, "Invalid price format: not a valid float", http.StatusBadRequest)
                    return
                }
            taxRateStr, ok := productData["taxRate"].(string)
                if !ok {
                    http.Error(w, "Invalid taxRate format: not a string", http.StatusBadRequest)
                    return
                }
        
                taxRate, err := strconv.ParseFloat(taxRateStr, 64)
                if err != nil {
                    http.Error(w, "Invalid taxRate format: not a valid float", http.StatusBadRequest)
                    return
                }
            taxStr, ok := productData["tax"].(string)
                if !ok {
                    http.Error(w, "Invalid tax format: not a string", http.StatusBadRequest)
                    return
                }
        
                tax, err := strconv.ParseFloat(taxStr, 64)
                if err != nil {
                    http.Error(w, "Invalid tax format: not a valid float", http.StatusBadRequest)
                    return
                }
                totalStr, ok := productData["total"].(string)
                if !ok {
                    http.Error(w, "Invalid total format: not a string", http.StatusBadRequest)
                    return
                }
        
                total, err := strconv.ParseFloat(totalStr, 64)
                if err != nil {
                    http.Error(w, "Invalid total format: not a valid float", http.StatusBadRequest)
                    return
                }
                // ... use quantity, price, taxRate, tax, total
            }           


            // Validate and convert data types
            saleDate, err := time.Parse("2006-01-02T15:04", saleDateStr) // ใช้ saleDateStr
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
                InvoiceNumber:  nextInvoiceNumber,
                CustomerName:   customerName,
                SaleDate:       saleDate,
                SalesPerson:    salesPerson,
                Description:    description,
                Note:           note,
                TaxID:          taxID,
                CustomerAddress: customerAddress,
                Discount:       discount,
            }

            if err := db.Create(&invoice).Error; err != nil {
                http.Error(w, "Error creating invoice", http.StatusInternalServerError)
                return
            }

                        // Extract product data
                        productIDs := req.Form["product_id"]
                        itemNames := req.Form["item_name"]
                        quantitiesStr := req.Form["quantity"]
                        pricesStr := req.Form["price"]
                        taxesStr := req.Form["tax"]

                        // Validate product data length
                        if len(productIDs) != len(itemNames) || len(productIDs) != len(quantitiesStr) || len(productIDs) != len(pricesStr) || len(productIDs) != len(taxesStr) {
                                http.Error(w, "Inconsistent product data", http.StatusBadRequest)
                                return
                        }

                        // Process each product
                        for i := range productIDs {
                                quantity, err := strconv.Atoi(quantitiesStr[i])
                                if err != nil {
                                        http.Error(w, fmt.Sprintf("Invalid quantity format for product %s", productIDs[i]), http.StatusBadRequest)
                                        return
                                }

                                price, err := strconv.ParseFloat(pricesStr[i], 64)
                                if err != nil {
                                        http.Error(w, fmt.Sprintf("Invalid price format for product %s", productIDs[i]), http.StatusBadRequest)
                                        return
                                }

                                taxRate, err := strconv.ParseFloat(taxesStr[i], 64)
                                if err != nil {
                                        http.Error(w, fmt.Sprintf("Invalid tax format for product %s", productIDs[i]), http.StatusBadRequest)
                                        return
                                }

                                // Validate tax_rate (0-100)
                                if taxRate < 0 || taxRate > 100 {
                                        http.Error(w, fmt.Sprintf("Invalid tax rate for product %s. Tax rate must be between 0 and 100.", productIDs[i]), http.StatusBadRequest)
                                        return
                                }

                                // Calculate tax and total for the item
                                tax := price * float64(quantity) * (taxRate / 100)
                                total := price*float64(quantity) + tax // Removed unused variable

                                // Check and create product if it doesn't exist
                                var product models.Product
                                if err := db.Where("id = ?", productIDs[i]).First(&product).Error; err != nil {
                                        if gorm.IsRecordNotFoundError(err) {
                                                product = models.Product{
                                                        ID:    productIDs[i],
                                                        Name:  itemNames[i],
                                                        Price: price,
                                                        Stock: 0,
                                                }
                                                if err := db.Create(&product).Error; err != nil {
                                                        http.Error(w, "Error creating product", http.StatusInternalServerError)
                                                        return
                                                }
                                        } else {
                                                http.Error(w, "Error finding product", http.StatusInternalServerError)
                                                return
                                        }
                                }

                                // Create Invoice Item
                                item := models.InvoiceItem{
                                        InvoiceID:  invoice.ID,
                                        ProductID:  productIDs[i],
                                        ItemName:   itemNames[i],
                                        Quantity:   quantity,
                                        Price:      price,
                                        TaxRate:    taxRate,
                                        Tax:        tax,
                                        Total:      total, // You might want to store this as well
                                }
                                if err := db.Create(&item).Error; err != nil {
                                        http.Error(w, "Error creating invoice item", http.StatusInternalServerError)
                                        return
                                }

                                // Update product stock
                                product.Stock -= quantity
                                if err := db.Save(&product).Error; err != nil {
                                        http.Error(w, "Error updating product stock", http.StatusInternalServerError)
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