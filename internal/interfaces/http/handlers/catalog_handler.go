package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	appcat "github.com/retechfin/retechfin-api/internal/application/catalog"
)

// CatalogHandler publica todos os enums de domínio com rótulo humano numa
// chamada (GET /api/v1/catalogs). Autenticado, sem RequireModule: o admin de
// qualquer módulo precisa dele. Ver docs/catalogs.md.
type CatalogHandler struct {
	svc *appcat.Service
}

func NewCatalogHandler(svc *appcat.Service) *CatalogHandler {
	return &CatalogHandler{svc: svc}
}

// catalogMaxAge é o Cache-Control do catálogo (1h); o conteúdo só muda em
// deploy, e o ETag revalida de graça.
const catalogMaxAge = "private, max-age=3600"

// All devolve { version, catalogs: { "<key>": [ {slug,name,description?} ] } }.
// ETag = version; If-None-Match igual → 304 sem corpo.
func (h *CatalogHandler) All(c *gin.Context) {
	etag := `"` + h.svc.Version() + `"`
	c.Header("ETag", etag)
	c.Header("Cache-Control", catalogMaxAge)
	if matchesETag(c.GetHeader("If-None-Match"), etag) {
		c.Status(http.StatusNotModified)
		return
	}
	c.JSON(http.StatusOK, h.svc.All())
}

// matchesETag aceita `*`, a tag exata ou uma lista separada por vírgula
// (RFC 9110 §13.1.2), com ou sem prefixo weak `W/`.
func matchesETag(ifNoneMatch, etag string) bool {
	if ifNoneMatch == "*" {
		return true
	}
	for _, raw := range strings.Split(ifNoneMatch, ",") {
		if strings.TrimPrefix(strings.TrimSpace(raw), "W/") == etag {
			return true
		}
	}
	return false
}
