package homesafety

import "github.com/retechfin/retechfin-api/internal/domain/catalog"

// Catálogos (enums) do módulo de segurança do lar publicados em
// GET /api/v1/catalogs (docs/catalogs.md). Não confundir com Catalog(), que é
// a lista de itens sugeridos para "adicionar do catálogo".

var categoryOrder = []Category{
	CategoryGas, CategoryAgua, CategoryIncendio, CategoryEletrica, CategoryClimatizacao, CategoryPragas,
	CategoryEstrutura, CategorySegurancaEletronica, CategoryPiscina, CategorySaude, CategoryOutros,
}

var categoryLabels = map[Category]string{
	CategoryGas:                 "Gás",
	CategoryAgua:                "Água",
	CategoryIncendio:            "Incêndio",
	CategoryEletrica:            "Elétrica",
	CategoryClimatizacao:        "Climatização",
	CategoryPragas:              "Pragas",
	CategoryEstrutura:           "Estrutura",
	CategorySegurancaEletronica: "Segurança eletrônica",
	CategoryPiscina:             "Piscina",
	CategorySaude:               "Saúde",
	CategoryOutros:              "Outros",
}

var riskTypeOrder = []RiskType{RiskFisico, RiskQuimico, RiskBiologico, RiskEletrico, RiskIncendio, RiskOutros}

var riskTypeLabels = map[RiskType]string{
	RiskFisico:    "Físico",
	RiskQuimico:   "Químico",
	RiskBiologico: "Biológico",
	RiskEletrico:  "Elétrico",
	RiskIncendio:  "Incêndio",
	RiskOutros:    "Outros",
}

var priorityOrder = []Priority{PriorityAlta, PriorityMedia, PriorityBaixa}

var priorityLabels = map[Priority]string{
	PriorityAlta:  "Alta",
	PriorityMedia: "Média",
	PriorityBaixa: "Baixa",
}

var eventTypeOrder = []EventType{EventManutencao, EventTroca, EventInstalacao, EventInspecao, EventRecarga, EventLimpeza}

var eventTypeLabels = map[EventType]string{
	EventManutencao: "Manutenção",
	EventTroca:      "Troca",
	EventInstalacao: "Instalação",
	EventInspecao:   "Inspeção",
	EventRecarga:    "Recarga",
	EventLimpeza:    "Limpeza",
}

// Status é calculado pela API a partir de next_due_date (não editável).
var statusOrder = []Status{StatusVencido, StatusAtencao, StatusProximo, StatusOK, StatusSemControle}

var statusLabels = map[Status]string{
	StatusVencido:     "Vencido",
	StatusAtencao:     "Atenção",
	StatusProximo:     "Próximo",
	StatusOK:          "Em dia",
	StatusSemControle: "Sem controle",
}

// Catalogs devolve os catálogos do módulo, na ordem de exibição.
func Catalogs() []catalog.Catalog {
	return []catalog.Catalog{
		{Key: "homesafety.category", Items: catalog.FromLabels(categoryOrder, categoryLabels)},
		{Key: "homesafety.risk_type", Items: catalog.FromLabels(riskTypeOrder, riskTypeLabels)},
		{Key: "homesafety.priority", Items: catalog.FromLabels(priorityOrder, priorityLabels)},
		{Key: "homesafety.event_type", Items: catalog.FromLabels(eventTypeOrder, eventTypeLabels)},
		{Key: "homesafety.status", Items: catalog.FromLabels(statusOrder, statusLabels)},
	}
}
