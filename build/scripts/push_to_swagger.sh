#!/usr/bin/env bash

# push the swagger api json to swagger hub
echo "pushing swagger.json to SwaggerHub"
VERSION=`jq -c '.info.version' httpapi/swagger.json -r`

curl -i -X POST \
  https://api.swaggerhub.com/apis/centrifuge.io/cent-node?version=${VERSION} \
  -H "Authorization: $SWAGGER_API_KEY" \
  -H "Content-Type: application/json" -d @./httpapi/swagger.json

exit $?
