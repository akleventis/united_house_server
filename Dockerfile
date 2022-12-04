FROM golang:latest

# set current working directory
WORKDIR $GOPATH/src/github.com/akleventis/united_house_server

# copy go.mod & go.sum to container
COPY go.* ./
RUN go mod download

# copy relevant folders over
# COPY .env ./.env
COPY lib/ ./lib/
COPY middleware/ ./middleware/
COPY uhp_api/ ./uhp_api/

# change working directory to properly build binary
WORKDIR $GOPATH/src/github.com/akleventis/united_house_server/uhp_api
RUN go build -o /uhp_api

EXPOSE 8080

# run executable
CMD [ "/uhp_api" ]

#  docker build --tag united_house .
# docker run --publish 8080:8080 united_house
