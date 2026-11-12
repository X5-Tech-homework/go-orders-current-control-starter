// Package docs содержит OpenAPI-описание и страницу Swagger UI.
package docs

import _ "embed"

//go:embed openapi.json
var OpenAPI []byte

const UI = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Orders API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>SwaggerUIBundle({url: "/swagger/openapi.json", dom_id: "#swagger-ui"});</script>
</body>
</html>`
