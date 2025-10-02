package controllers

import "github.com/gin-gonic/gin"

func RapiDoc(c *gin.Context) {
	rapidocHTML := `<!doctype html>
<html>
  <head>
    <meta charset="utf-8"/>
    <title>Currency Converter API – RapiDoc</title>
    <meta name="viewport" content="width=device-width, initial-scale=1"/>
    <script type="module" src="https://unpkg.com/rapidoc/dist/rapidoc-min.js"></script>
    <style>
      html, body { height: 100%; margin: 0; }
      rapi-doc { height: 100vh; }
    </style>
  </head>
  <body>
    <rapi-doc
      id="rd"
      spec-url="/openapi.json"
      server-url="http://localhost:8080/api/v1"
      render-style="read"
      show-header="true"
      theme="light"
      nav-item-spacing="compact"
      allow-try="true"
      show-curl-before-try="true"
      schema-style="table"
      sort-endpoints-by="path"
      sort-tags="true"
      use-path-in-nav-bar="true"
      allow-authentication="true"
    >
      <div slot="nav-logo" style="font-weight:700;padding:8px 12px">Currency Converter API</div>
      <div slot="header" style="padding:8px 12px">Currency Converter API</div>
    </rapi-doc>
  </body>
</html>`
	c.Data(200, "text/html; charset=utf-8", []byte(rapidocHTML))
}
