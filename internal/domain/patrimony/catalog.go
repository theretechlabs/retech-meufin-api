package patrimony

import "github.com/retechfin/retechfin-api/internal/domain/catalog"

// Catálogos do módulo de patrimônio publicados em GET /api/v1/catalogs
// (docs/catalogs.md). Novo valor num enum = nova constante + rótulo aqui.

var propertyTypeOrder = []PropertyType{
	PropertyCasa, PropertyApartamento, PropertyTerreno, PropertyComercial, PropertyRural, PropertyOutros,
}

var propertyTypeLabels = map[PropertyType]string{
	PropertyCasa:        "Casa",
	PropertyApartamento: "Apartamento",
	PropertyTerreno:     "Terreno",
	PropertyComercial:   "Comercial",
	PropertyRural:       "Rural",
	PropertyOutros:      "Outros",
}

var propertyDocTypeOrder = []PropertyDocType{
	DocEscritura, DocMatricula, DocIPTU, DocContrato, DocSeguro, DocPlanta, DocOutros,
}

var propertyDocTypeLabels = map[PropertyDocType]string{
	DocEscritura: "Escritura",
	DocMatricula: "Matrícula",
	DocIPTU:      "Carnê de IPTU",
	DocContrato:  "Contrato",
	DocSeguro:    "Apólice de seguro",
	DocPlanta:    "Planta",
	DocOutros:    "Outros",
}

var assetTypeOrder = []AssetType{AssetProperty, AssetVehicle}

var assetTypeLabels = map[AssetType]string{
	AssetProperty: "Imóvel",
	AssetVehicle:  "Veículo",
}

var taxTypeOrder = []TaxType{
	TaxIPTU, TaxIPVA, TaxLicenciamento, TaxDPVAT, TaxCondominio, TaxSeguroPredial,
	TaxTaxaLixo, TaxTaxaBombeiros, TaxOutros,
}

var taxTypeLabels = map[TaxType]string{
	TaxIPTU:          "IPTU",
	TaxIPVA:          "IPVA",
	TaxLicenciamento: "Licenciamento",
	TaxDPVAT:         "DPVAT",
	TaxCondominio:    "Condomínio",
	TaxSeguroPredial: "Seguro predial",
	TaxTaxaLixo:      "Taxa de lixo",
	TaxTaxaBombeiros: "Taxa de bombeiros",
	TaxOutros:        "Outros",
}

var taxStatusOrder = []TaxStatus{TaxStatusPending, TaxStatusPaid, TaxStatusOverdue, TaxStatusPartial}

var taxStatusLabels = map[TaxStatus]string{
	TaxStatusPending: "Pendente",
	TaxStatusPaid:    "Pago",
	TaxStatusOverdue: "Vencido",
	TaxStatusPartial: "Parcial",
}

// Catalogs devolve os catálogos do módulo de patrimônio, na ordem de exibição.
func Catalogs() []catalog.Catalog {
	return []catalog.Catalog{
		{Key: "patrimony.property_type", Items: catalog.FromLabels(propertyTypeOrder, propertyTypeLabels)},
		{Key: "patrimony.property_doc_type", Items: catalog.FromLabels(propertyDocTypeOrder, propertyDocTypeLabels)},
		{Key: "patrimony.asset_type", Items: catalog.FromLabels(assetTypeOrder, assetTypeLabels)},
		{Key: "patrimony.tax_type", Items: catalog.FromLabels(taxTypeOrder, taxTypeLabels)},
		{Key: "patrimony.tax_status", Items: catalog.FromLabels(taxStatusOrder, taxStatusLabels)},
	}
}
