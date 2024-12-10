package main

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Category    string  `json:"category"`
}

var (
	products = make(map[string]Product)
	mutex    = sync.RWMutex{}
)

func main() {
	http.HandleFunc("/products", handleProducts)
	http.HandleFunc("/products/", handleProductByID)

	http.ListenAndServe(":8080", nil)
}

func handleProducts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mutex.RLock()
		defer mutex.RUnlock()

		var productList []Product
		for _, product := range products {
			productList = append(productList, product)
		}
		json.NewEncoder(w).Encode(productList)

	case http.MethodPost:
		var product Product
		if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		product.ID = uuid.New().String()
		mutex.Lock()
		products[product.ID] = product
		mutex.Unlock()

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(product)
	}
}

func handleProductByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/products/"):]

	mutex.RLock()
	product, exists := products[id]
	mutex.RUnlock()

	if !exists {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(product)

	case http.MethodPut:
		var updatedProduct Product
		if err := json.NewDecoder(r.Body).Decode(&updatedProduct); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		updatedProduct.ID = id
		mutex.Lock()
		products[id] = updatedProduct
		mutex.Unlock()

		json.NewEncoder(w).Encode(updatedProduct)

	case http.MethodDelete:
		mutex.Lock()
		delete(products, id)
		mutex.Unlock()

		w.WriteHeader(http.StatusNoContent)
	}
}
