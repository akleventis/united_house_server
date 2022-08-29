module uhp_api.go

go 1.17

require github.com/lib/pq v1.10.4

require (
	github.com/mailjet/mailjet-apiv3-go v0.0.0-20201009050126-c24bc15a9394 // indirect
	golang.org/x/sys v0.0.0-20220519141025-dcacdad47464 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
)

require (
	github.com/akleventis/united_house_server v0.0.0-20220617061045-083594f93391
	github.com/gorilla/mux v1.8.0
	github.com/joho/godotenv v1.4.0
	github.com/rs/cors v1.8.2
	github.com/sirupsen/logrus v1.8.1
	github.com/stripe/stripe-go v70.15.0+incompatible
	github.com/stripe/stripe-go/v72 v72.95.0
	golang.org/x/time v0.0.0-20220224211638-0e9765cccd65
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
)

replace github.com/akleventis/united_house_server => ../united_house_go
