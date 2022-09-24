package main

import (
	"net/http"
	"os"

	m "github.com/akleventis/united_house_server/middleware"

	// TODO: figure out why imports only work with named variable
	artists "github.com/akleventis/united_house_server/uhp_api/handlers/artists"
	auth "github.com/akleventis/united_house_server/uhp_api/handlers/auth"
	checkout "github.com/akleventis/united_house_server/uhp_api/handlers/checkout"
	email "github.com/akleventis/united_house_server/uhp_api/handlers/email"
	events "github.com/akleventis/united_house_server/uhp_api/handlers/events"
	images "github.com/akleventis/united_house_server/uhp_api/handlers/images"
	products "github.com/akleventis/united_house_server/uhp_api/handlers/products"
	"github.com/akleventis/united_house_server/uhp_db"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mailjet/mailjet-apiv3-go"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	stripe "github.com/stripe/stripe-go"
)

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := uhp_db.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()

	stripe.Key = os.Getenv("STRIPE_KEY")

	awsRegion := os.Getenv("AWS_REGION")
	awsAccess := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecret := os.Getenv("AWS_SECRET_ACCESS_KEY")
	s3Bucket := os.Getenv("AWS_BUCKET")

	s3Config := &aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(awsAccess, awsSecret, ""),
	}

	s3Session, err := aws_session.NewSession(s3Config)
	if err != nil {
		log.Fatal(err)
	}

	s3Client := s3.New(s3Session)

	router := mux.NewRouter()

	// stripe
	checkout := checkout.NewHandler(db)
	router.HandleFunc("/checkout", m.Limit(checkout.HandleCheckout(), m.RL10)).Methods("POST")
	router.HandleFunc("/webhook", m.Limit(checkout.HandleWebhook(), m.RL10)).Methods("POST")

	// products
	products := products.NewHandler(db)
	router.HandleFunc("/products", m.Limit(products.GetProducts(), m.RL50)).Methods("GET")
	router.HandleFunc("/product/{id}", m.Limit(m.Auth(products.GetProduct()), m.RL30)).Methods("GET")       // admin
	router.HandleFunc("/product", m.Limit(m.Auth(products.CreateProduct()), m.RL30)).Methods("POST")        // admin
	router.HandleFunc("/product/{id}", m.Limit(m.Auth(products.UpdateProduct()), m.RL30)).Methods("PATCH")  // admin
	router.HandleFunc("/product/{id}", m.Limit(m.Auth(products.DeleteProduct()), m.RL30)).Methods("DELETE") // admin

	// events
	events := events.NewHandler(db)
	router.HandleFunc("/events", m.Limit(events.GetEvents(), m.RL50)).Methods("GET")
	router.HandleFunc("/event/{id}", m.Limit(events.GetEvent(), m.RL50)).Methods("GET")
	router.HandleFunc("/event", m.Limit(m.Auth(events.CreateEvent()), m.RL30)).Methods("POST")        // admin
	router.HandleFunc("/event/{id}", m.Limit(m.Auth(events.UpdateEvent()), m.RL30)).Methods("PATCH")  // admin
	router.HandleFunc("/event/{id}", m.Limit(m.Auth(events.DeleteEvent()), m.RL30)).Methods("DELETE") // admin

	// featured artist soundcloud iframe
	artists := artists.NewHandler(db)
	router.HandleFunc("/artist", m.Limit(artists.GetArtists(), m.RL50)).Methods("GET")
	router.HandleFunc("/artist", m.Limit(m.Auth(artists.CreateArtist()), m.RL30)).Methods("POST")         // admin
	router.HandleFunc("/artists", m.Limit(m.Auth(artists.GetArtist()), m.RL50)).Methods("GET")            // admin
	router.HandleFunc("/artists/{id}", m.Limit(m.Auth(artists.UpdateArtist()), m.RL30)).Methods("PATCH")  // admin
	router.HandleFunc("/artists/{id}", m.Limit(m.Auth(artists.DeleteArtist()), m.RL30)).Methods("DELETE") // admin

	router.HandleFunc("/featured_artists", m.Limit(artists.GetFeaturedArtists(), m.RL50)).Methods("GET")
	router.HandleFunc("/featured_artist", m.Limit(m.Auth(artists.CreateFeaturedArtist()), m.RL30)).Methods("POST")        // admin
	router.HandleFunc("/featured_artist/{id}", m.Limit(m.Auth(artists.UpdateFeaturedArtist()), m.RL30)).Methods("PATCH")  // admin
	router.HandleFunc("/featured_artist/{id}", m.Limit(m.Auth(artists.DeleteFeaturedArtist()), m.RL30)).Methods("DELETE") // admin

	// email
	mailjetClient := mailjet.NewMailjetClient(os.Getenv("MAILJET_KEY"), os.Getenv("MAILJET_SECRET"))
	email := email.NewHandler(mailjetClient)
	router.HandleFunc("/mail", m.Limit(email.SendEmail(), m.RL5)).Methods("POST")

	// aws s3 image upload
	images := images.NewHandler(db, s3Client, s3Session, s3Bucket)
	router.HandleFunc("/images", m.Limit(images.GetImage(), m.RL30)).Methods("GET")
	router.HandleFunc("/images", m.Limit(m.Auth(images.UploadImage()), m.RL10)).Methods("POST")   // admin
	router.HandleFunc("/images", m.Limit(m.Auth(images.DeleteImage()), m.RL10)).Methods("DELETE") // admin

	// admin sign-in
	auth := auth.NewHandler(db)
	router.HandleFunc("/auth", auth.SignIn()).Methods("POST")

	handler := cors.AllowAll().Handler(router)
	http.ListenAndServe(":5001", handler)
}
