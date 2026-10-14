package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	appcat "github.com/retechfin/retechfin-api/internal/application/catalog"
)

func newCatalogRouter(t *testing.T) (*gin.Engine, *appcat.Service) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	svc, err := appcat.New()
	if err != nil {
		t.Fatalf("catalog service: %v", err)
	}
	r := gin.New()
	r.GET("/api/v1/catalogs", NewCatalogHandler(svc).All)
	return r, svc
}

func TestCatalogHandler_All_200WithETag(t *testing.T) {
	r, svc := newCatalogRouter(t)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalogs", nil)
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d, esperado 200; body=%s", rec.Code, rec.Body.String())
	}
	wantETag := `"` + svc.Version() + `"`
	if got := rec.Header().Get("ETag"); got != wantETag {
		t.Errorf("ETag %q, esperado %q", got, wantETag)
	}
	if got := rec.Header().Get("Cache-Control"); got != "private, max-age=3600" {
		t.Errorf("Cache-Control %q", got)
	}

	var body struct {
		Version  string `json:"version"`
		Catalogs map[string][]struct {
			Slug        string  `json:"slug"`
			Name        string  `json:"name"`
			Description *string `json:"description"`
		} `json:"catalogs"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body.Version != svc.Version() {
		t.Errorf("version %q != %q", body.Version, svc.Version())
	}
	pm, ok := body.Catalogs["finance.payment_method"]
	if !ok || len(pm) == 0 {
		t.Fatalf("finance.payment_method ausente no corpo")
	}
	if pm[0].Slug != "pix" || pm[0].Name != "Pix" {
		t.Errorf("primeiro item de payment_method = %+v, esperado pix/Pix", pm[0])
	}
	// description é omitida quando vazia (omitempty), nunca "".
	for key, items := range body.Catalogs {
		for _, it := range items {
			if it.Description != nil && *it.Description == "" {
				t.Errorf("%s/%s: description vazia serializada; deveria ser omitida", key, it.Slug)
			}
		}
	}
}

func TestCatalogHandler_All_304OnIfNoneMatch(t *testing.T) {
	r, svc := newCatalogRouter(t)
	etag := `"` + svc.Version() + `"`

	cases := map[string]string{
		"exata":      etag,
		"weak":       "W/" + etag,
		"lista":      `"outra", ` + etag,
		"wildcard":   "*",
		"weak+lista": `W/"x", W/` + etag,
	}
	for name, inm := range cases {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/catalogs", nil)
		req.Header.Set("If-None-Match", inm)
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotModified {
			t.Errorf("%s: status %d, esperado 304", name, rec.Code)
		}
		if rec.Body.Len() != 0 {
			t.Errorf("%s: 304 com corpo (%d bytes)", name, rec.Body.Len())
		}
		if got := rec.Header().Get("ETag"); got != etag {
			t.Errorf("%s: ETag %q no 304, esperado %q", name, got, etag)
		}
	}

	// Tag diferente → conteúdo completo.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/catalogs", nil)
	req.Header.Set("If-None-Match", `"desatualizada"`)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("tag desatualizada: status %d, esperado 200", rec.Code)
	}
}
