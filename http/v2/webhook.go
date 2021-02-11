package v2

import "net/http"

// Webhook is a place holder to describe webhook response in swagger
// @summary Webhook is a place holder to describe webhook response in swagger.
// @description Webhook is a place holder to describe webhook response in swagger.
// @id webhook
// @tags Webhook
// @accept json
// @produce json
// @success 200 {object} notification.Message
// @router  /webhook [post]
func Webhook(http.ResponseWriter, *http.Request) {}
