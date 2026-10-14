// Package catalog define o contrato dos catálogos de enum publicados pela API
// (GET /api/v1/catalogs). É um pacote folha: não importa nenhum módulo; cada
// módulo (`finance`, `health`, `vehicle`…) expõe `Catalogs()` devolvendo os
// seus próprios catálogos com estes tipos. Ver docs/catalogs.md.
package catalog

// Item é um valor de enum com rótulo humano. O `slug` é o que o admin envia e
// o banco grava; `name`/`description` são só apresentação.
type Item struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// Catalog é a lista ordenada de itens de um enum, identificada pela chave
// `<módulo>.<enum>` (ex.: `finance.payment_method`) — a mesma chave do
// `CatalogKey` do admin.
type Catalog struct {
	Key   string
	Items []Item
}

// FromLabels monta os itens de um enum tipado na ordem informada, pegando o
// rótulo no mapa. Slug sem rótulo gera Item com Name vazio, que o serviço de
// catálogo rejeita na construção — assim a falta de rótulo quebra o teste em
// vez de chegar ao admin.
func FromLabels[T ~string](order []T, labels map[T]string) []Item {
	items := make([]Item, 0, len(order))
	for _, v := range order {
		items = append(items, Item{Slug: string(v), Name: labels[v]})
	}
	return items
}
