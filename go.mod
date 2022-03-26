module uhp_api.go

go 1.17

require github.com/lib/pq v1.10.4

require golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd // indirect

require (
	github.com/akleventis/united_house_server v0.0.0-20220324025155-7a5ffcda5065
	github.com/joho/godotenv v1.4.0
	github.com/rs/cors v1.8.2
	github.com/sirupsen/logrus v1.8.1
	github.com/stripe/stripe-go/v72 v72.95.0
)

replace github.com/akleventis/united_house_server => ../united_house_go
