start:
	go run .

test-integration:
	GIN_MODE=test go test -v ./tests/...

start-db:
	mongod --replSet rs0 --port 27017 --oplogSize 128 --fork --dbpath /data/db --logpath ./mongod.log
	mongod --replSet rs0 --port 27018 --oplogSize 128 --fork --dbpath /data/db1 --logpath ./mongod1.log
	mongod --replSet rs0 --port 27019 --oplogSize 128 --fork --dbpath /data/db2 --logpath ./mongod2.log
