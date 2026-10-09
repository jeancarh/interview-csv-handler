package httpapi

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"strings"

	"github.com/jeancarh/interview-csv-handler/backend/internal/catalog"
	"github.com/jeancarh/interview-csv-handler/backend/internal/checkout"
)

type Deps struct {
	Catalog        *catalog.Service
	Importer       *catalog.Importer
	Checkout       *checkout.Service
	Ping           func(ctx context.Context) error
	Assets         fs.FS
	Spec           []byte
	MaxUploadBytes int64
	MaxImportRows  int
}

type api struct {
	catalog        *catalog.Service
	importer       *catalog.Importer
	checkout       *checkout.Service
	ping           func(ctx context.Context) error
	maxUploadBytes int64
	maxImportRows  int
}

func NewRouter(d Deps) http.Handler {
	a := &api{
		catalog:        d.Catalog,
		importer:       d.Importer,
		checkout:       d.Checkout,
		ping:           d.Ping,
		maxUploadBytes: d.MaxUploadBytes,
		maxImportRows:  d.MaxImportRows,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.health)

	mux.HandleFunc("GET /api/v1/products", a.searchProducts)
	mux.HandleFunc("POST /api/v1/products", a.createProduct)
	mux.HandleFunc("GET /api/v1/products/{id}", a.getProduct)
	mux.HandleFunc("PUT /api/v1/products/{id}", a.updateProduct)
	mux.HandleFunc("DELETE /api/v1/products/{id}", a.deleteProduct)
	mux.HandleFunc("POST /api/v1/products/import", a.importProducts)
	mux.HandleFunc("GET /api/v1/categories", a.listCategories)
	mux.HandleFunc("POST /api/v1/orders", a.placeOrder)
	mux.HandleFunc("GET /api/v1/orders/{id}", a.getOrder)

	mux.HandleFunc("GET /api/openapi.yaml", serveSpec(d.Spec))
	mux.HandleFunc("GET /docs", serveDocs)
	mux.Handle("/", fallback(spa(d.Assets)))

	return withRequestID(withLogging(withRecover(mux)))
}

func serveSpec(spec []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(spec)
	}
}

const docsPage = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Shop API docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css">
</head>
<body>
  <div id="swagger"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
  <script>SwaggerUIBundle({ url: "/api/openapi.yaml", dom_id: "#swagger" });</script>
</body>
</html>`

func serveDocs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(docsPage))
}

func fallback(ui http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeProblem(w, http.StatusNotFound, "no such endpoint", nil)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			writeProblem(w, http.StatusMethodNotAllowed, "", nil)
			return
		}
		ui.ServeHTTP(w, r)
	})
}

func spa(assets fs.FS) http.Handler {
	files := http.FileServerFS(assets)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name != "" {
			if info, err := fs.Stat(assets, name); err == nil && !info.IsDir() {
				if strings.HasPrefix(name, "assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				}
				files.ServeHTTP(w, r)
				return
			}
		}
		index, err := fs.ReadFile(assets, "index.html")
		if errors.Is(err, fs.ErrNotExist) {
			http.Error(w, "The UI is not built into this binary. Use `make dev` or `docker compose up` instead.", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(index)
	})
}
