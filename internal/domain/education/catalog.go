package education

import "github.com/retechfin/retechfin-api/internal/domain/catalog"

// Catálogos do módulo de educação publicados em GET /api/v1/catalogs
// (docs/catalogs.md). Novo valor num enum = nova constante + rótulo aqui.

var stageOrder = []Stage{
	StageBercario, StageInfantil, StageFundamental1, StageFundamental2, StageMedio,
	StageTecnico, StagePreVestibular, StageSuperior, StagePos,
}

var stageLabels = map[Stage]string{
	StageBercario:      "Berçário",
	StageInfantil:      "Educação Infantil",
	StageFundamental1:  "Fundamental I",
	StageFundamental2:  "Fundamental II",
	StageMedio:         "Ensino Médio",
	StageTecnico:       "Técnico",
	StagePreVestibular: "Pré-vestibular",
	StageSuperior:      "Ensino Superior",
	StagePos:           "Pós-graduação",
}

var shiftOrder = []Shift{ShiftManha, ShiftTarde, ShiftIntegral, ShiftNoite, ShiftEAD}

var shiftLabels = map[Shift]string{
	ShiftManha:    "Manhã",
	ShiftTarde:    "Tarde",
	ShiftIntegral: "Integral",
	ShiftNoite:    "Noite",
	ShiftEAD:      "EAD",
}

var listStatusOrder = []ListStatus{ListStatusPlanejada, ListStatusEmCompra, ListStatusConcluida}

var listStatusLabels = map[ListStatus]string{
	ListStatusPlanejada: "Planejada",
	ListStatusEmCompra:  "Em compra",
	ListStatusConcluida: "Concluída",
}

var itemCategoryOrder = []ItemCategory{
	ItemCategoryPapelaria, ItemCategoryLivros, ItemCategoryUniforme, ItemCategoryMochila,
	ItemCategoryEletronicos, ItemCategoryArte, ItemCategoryHigiene, ItemCategoryOutros,
}

var itemCategoryLabels = map[ItemCategory]string{
	ItemCategoryPapelaria:   "Papelaria",
	ItemCategoryLivros:      "Livros",
	ItemCategoryUniforme:    "Uniforme",
	ItemCategoryMochila:     "Mochila",
	ItemCategoryEletronicos: "Eletrônicos",
	ItemCategoryArte:        "Arte",
	ItemCategoryHigiene:     "Higiene",
	ItemCategoryOutros:      "Outros",
}

// Catalogs devolve os catálogos do módulo de educação, na ordem de exibição.
func Catalogs() []catalog.Catalog {
	return []catalog.Catalog{
		{Key: "education.stage", Items: catalog.FromLabels(stageOrder, stageLabels)},
		{Key: "education.shift", Items: catalog.FromLabels(shiftOrder, shiftLabels)},
		{Key: "education.list_status", Items: catalog.FromLabels(listStatusOrder, listStatusLabels)},
		{Key: "education.item_category", Items: catalog.FromLabels(itemCategoryOrder, itemCategoryLabels)},
	}
}
