#!/bin/sh
SWAGGER_PATH=$GOPATH/src/github.com/centrifuge/go-centrifuge/centrifuge/invoice

echo "Launching swagger-ui docker container..."
echo "Loading swagger.json from: $SWAGGER_PATH"
echo "Go to http://localhost:8085/ to access it once it's done starting."
docker run --rm -p 8085:8080 -e SWAGGER_JSON=/data/invoice.swagger.json -v $SWAGGER_PATH:/data swaggerapi/swagger-ui

