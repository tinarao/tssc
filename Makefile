APP_NAME = "tssc"

run: build 
	./bin/${APP_NAME}

build:
	go build -o ./bin/${APP_NAME}
