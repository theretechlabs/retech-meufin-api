// Package catalog monta a resposta de GET /api/v1/catalogs: todos os enums de
// domínio com rótulo humano, numa chamada, para o admin não manter constantes
// `*_OPTIONS`/`*_LABEL`. Ver docs/catalogs.md.
package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	dom "github.com/retechfin/retechfin-api/internal/domain/catalog"
	domedu "github.com/retechfin/retechfin-api/internal/domain/education"
	domfin "github.com/retechfin/retechfin-api/internal/domain/finance"
	domhealth "github.com/retechfin/retechfin-api/internal/domain/health"
	domhs "github.com/retechfin/retechfin-api/internal/domain/homesafety"
	dompat "github.com/retechfin/retechfin-api/internal/domain/patrimony"
	domveh "github.com/retechfin/retechfin-api/internal/domain/vehicle"
	domwar "github.com/retechfin/retechfin-api/internal/domain/warranty"
)

// Response é o corpo de GET /api/v1/catalogs. `Version` é o hash do conteúdo
// (também vai no ETag): muda quando qualquer catálogo muda.
type Response struct {
	Version  string                `json:"version"`
	Catalogs map[string][]dom.Item `json:"catalogs"`
}

// Service serve os catálogos. O conteúdo é estático (código Go), então a
// resposta é montada e validada uma vez, em New.
type Service struct {
	resp Response
}

// registry concatena os catálogos de cada módulo. Módulo novo = uma linha aqui.
func registry() []dom.Catalog {
	var all []dom.Catalog
	all = append(all, domfin.Catalogs()...)
	all = append(all, domedu.Catalogs()...)
	all = append(all, domhealth.Catalogs()...)
	all = append(all, domhs.Catalogs()...)
	all = append(all, dompat.Catalogs()...)
	all = append(all, domveh.Catalogs()...)
	all = append(all, domwar.Catalogs()...)
	return all
}

// New monta a resposta e falha na partida se algum catálogo estiver quebrado
// (chave duplicada, catálogo vazio, slug/nome vazio, slug duplicado) — melhor
// a API não subir do que o admin receber um select sem rótulo.
func New() (*Service, error) {
	cats := registry()
	byKey := make(map[string][]dom.Item, len(cats))
	for _, c := range cats {
		if c.Key == "" {
			return nil, fmt.Errorf("catalog: catálogo sem chave")
		}
		if _, dup := byKey[c.Key]; dup {
			return nil, fmt.Errorf("catalog %q: chave duplicada", c.Key)
		}
		if len(c.Items) == 0 {
			return nil, fmt.Errorf("catalog %q: sem itens", c.Key)
		}
		seen := make(map[string]struct{}, len(c.Items))
		for i, it := range c.Items {
			if it.Slug == "" {
				return nil, fmt.Errorf("catalog %q: item %d sem slug", c.Key, i)
			}
			if it.Name == "" {
				return nil, fmt.Errorf("catalog %q: slug %q sem rótulo", c.Key, it.Slug)
			}
			if _, dup := seen[it.Slug]; dup {
				return nil, fmt.Errorf("catalog %q: slug %q duplicado", c.Key, it.Slug)
			}
			seen[it.Slug] = struct{}{}
		}
		byKey[c.Key] = c.Items
	}
	version, err := computeVersion(cats)
	if err != nil {
		return nil, err
	}
	return &Service{resp: Response{Version: version, Catalogs: byKey}}, nil
}

// MustNew é New para a composição em main/testes: catálogo quebrado é bug de
// programação, não condição de runtime.
func MustNew() *Service {
	s, err := New()
	if err != nil {
		panic(err)
	}
	return s
}

// All devolve a resposta montada em New.
func (s *Service) All() Response { return s.resp }

// Version é o hash do conteúdo (ETag).
func (s *Service) Version() string { return s.resp.Version }

// computeVersion é sha256 do JSON canônico dos catálogos (slice ordenado pelo
// registro, não o map), truncado em 16 hex — determinístico entre processos.
func computeVersion(cats []dom.Catalog) (string, error) {
	type entry struct {
		Key   string     `json:"key"`
		Items []dom.Item `json:"items"`
	}
	canon := make([]entry, 0, len(cats))
	for _, c := range cats {
		canon = append(canon, entry{Key: c.Key, Items: c.Items})
	}
	b, err := json.Marshal(canon)
	if err != nil {
		return "", fmt.Errorf("catalog: serializar versão: %w", err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])[:16], nil
}
