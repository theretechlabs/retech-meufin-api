package warranty

import "github.com/retechfin/retechfin-api/internal/domain/catalog"

// Catálogos do módulo de garantias publicados em GET /api/v1/catalogs
// (docs/catalogs.md). Novo valor num enum = nova constante + rótulo aqui.

var categoryOrder = []Category{
	CategoryEletrodomestico, CategoryEletronico, CategoryInformatica, CategoryCelular, CategoryMovel,
	CategoryVeiculo, CategoryImovel, CategoryFerramenta, CategoryBrinquedo, CategoryEsporte, CategoryOutros,
}

var categoryLabels = map[Category]string{
	CategoryEletrodomestico: "Eletrodoméstico",
	CategoryEletronico:      "Eletrônico",
	CategoryInformatica:     "Informática",
	CategoryCelular:         "Celular",
	CategoryMovel:           "Móvel",
	CategoryVeiculo:         "Veículo",
	CategoryImovel:          "Imóvel",
	CategoryFerramenta:      "Ferramenta",
	CategoryBrinquedo:       "Brinquedo",
	CategoryEsporte:         "Esporte",
	CategoryOutros:          "Outros",
}

// Status é calculado pela API a partir da data de expiração (não editável).
var statusOrder = []Status{StatusVigente, StatusExpiraEmBreve, StatusExpirada}

var statusLabels = map[Status]string{
	StatusVigente:       "Vigente",
	StatusExpiraEmBreve: "Expira em breve",
	StatusExpirada:      "Expirada",
}

var docTypeOrder = []DocType{DocNotaFiscal, DocCertificado, DocGarantiaEstendida, DocManual, DocOutros}

var docTypeLabels = map[DocType]string{
	DocNotaFiscal:        "Nota fiscal",
	DocCertificado:       "Certificado de garantia",
	DocGarantiaEstendida: "Garantia estendida",
	DocManual:            "Manual",
	DocOutros:            "Outros",
}

// Catalogs devolve os catálogos do módulo de garantias, na ordem de exibição.
func Catalogs() []catalog.Catalog {
	return []catalog.Catalog{
		{Key: "warranty.category", Items: catalog.FromLabels(categoryOrder, categoryLabels)},
		{Key: "warranty.status", Items: catalog.FromLabels(statusOrder, statusLabels)},
		{Key: "warranty.doc_type", Items: catalog.FromLabels(docTypeOrder, docTypeLabels)},
	}
}
