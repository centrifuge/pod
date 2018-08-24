#!/usr/bin/env bash

# push the swagger api json to swagger hub
echo "pushing swagger.json to SwaggerHub"

curl -i -X POST \
  http://api.swaggerhub.com/apis/centrifuge.io/cent-node?version=0.0.1 \
  -H "Authorization: $SWAGGER_API_KEY" \
  -H "Content-Type: application/json" -d @./gen/swagger.json

exit $?