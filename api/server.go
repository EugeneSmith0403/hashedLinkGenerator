package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

// RegisterDocsRoutes adds /swagger/ UI and /api/openapi.yaml to the mux.
func RegisterDocsRoutes(mux *http.ServeMux, specPath string) {
	mux.HandleFunc("GET /api/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		abs, err := filepath.Abs(specPath)
		if err != nil {
			http.Error(w, "spec not found", http.StatusInternalServerError)
			return
		}
		data, err := os.ReadFile(abs)
		if err != nil {
			http.Error(w, "spec not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
		w.Write(data)
	})

	mux.HandleFunc("GET /swagger/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, swaggerUIHTML)
	})
}

const swaggerUIHTML = `<!DOCTYPE html>
<html>
<head>
  <title>API Docs</title>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
<script>
  SwaggerUIBundle({
    url: "/api/openapi.yaml",
    dom_id: '#swagger-ui',
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
    layout: "BaseLayout"
  })
</script>
</body>
</html>`
