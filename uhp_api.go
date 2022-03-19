package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	stripe "github.com/stripe/stripe-go"
)

// server struct (https://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years.html)
type server struct {
	// db *someDatabase
	router *http.ServeMux
}

// make prodcut struct
type product struct {
	id    int
	price int
	name  string
}

// getProductsDB() (map[int]product, err) {
//
// }
func getProducts() map[int]product {
	// init test product
	p1 := product{
		id:    1,
		price: 25,
		name:  "T-Shirt (s)",
	}
	p2 := product{
		id:    2,
		price: 25,
		name:  "T-Shirt (m)",
	}
	p3 := product{
		id:    3,
		price: 15,
		name:  "Bucket Hat",
	}
	products := make(map[int]product)
	products[1] = p1
	products[2] = p2
	products[3] = p3
	return products
}

// handle get post request
func (s *server) handleCheckout() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		products := getProducts()
		fmt.Println(products)
	}
}

// make server into http.Handler
// func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	s.router.ServeHTTP(w, r)
// }

// func main() {
// 	if err := run(); err != nil {
// 		fmt.Fprintf(os.Stderr, "%s\n", err)
// 		os.Exit(1)
// 	}
// }

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	stripe.Key = os.Getenv("STRIPE_KEY")

	s := &server{
		// database
		router: http.NewServeMux(),
	}

	s.router.HandleFunc("/checkout", s.handleCheckout())

	http.ListenAndServe(":5001", s.router)
}
