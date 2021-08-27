# GO-INSTAGRAM-API

A simple rest api designed to function like instagram

## TECH STACK

- Golang (Gin Framework)
- Docker
- MongoDB

## SETUP

### USING DOCKER

- Run `docker-compose -f docker-compose.mongo.yml up -d` to start three mongod instances in detached mode
- Connect to one of the instances e.g `docker exec -it mongo1 mongo`
- Create a config for the replica set with the code below

```javascript
var config = {
  _id: "rs0",
  members: [
    {
      _id: 0,
      host: "mongodb:27017",
    },
    {
      _id: 1,
      host: "mongodb1:27017",
    },
    {
      _id: 2,
      host: "mongodb2:27017",
    },
  ],
};
```

- Initiate the replica set `rs.initiate(config)`
- View the replica set `rs.conf()`
- Verify that the replica set has a primary. `rs.status()`
- Create a .env.prod file and add values for the following environmental variables
  ` APP_ACCESS_SECRET, AWS_ACCESS_KEY_ID, AWS_BUCKET, AWS_DEFAULT_REGION AWS_SECRET_ACCESS_KEY, CLIENT_ORIGIN, GIN_MODE=release, MONGODB_URI=mongodb://mongo1:27017,mongo2:27017,mongo3:27017/?replicaSet=rs0, MONGODB_NAME`
- Run `docker-compose -f docker-compose.backend.yml up -d` to start the API
- Run `docker-compose -f docker-compose.mongo.yml -f docker-compose.backend.yml down` to stop all services

### WITHOUT DOCKER

- Ensure you have mongodb installed locally
- Run `mkdir -p /data/db /data/db1 /data/db2` to create these three directories for each mongod instance
- Run `make start-db` to start all instances
- Configure the replica-set using the docker setup above but using `localhost:27017,localhost:27018,localhost:27019` as member hosts
- Configure environmental variables as above
- Run `make start` to start application

## TESTING

- Setup a replica-set with or without docker or even remote.
- Run `make test-integration` for integration tests
