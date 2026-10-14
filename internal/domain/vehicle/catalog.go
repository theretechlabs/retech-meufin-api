package vehicle

import "github.com/retechfin/retechfin-api/internal/domain/catalog"

// Catálogos do módulo de frota publicados em GET /api/v1/catalogs
// (docs/catalogs.md). Novo valor num enum = nova constante + rótulo aqui.

var fuelTypeOrder = []FuelType{FuelFlex, FuelGasolina, FuelEtanol, FuelDiesel, FuelEletrico, FuelHibrido}

var fuelTypeLabels = map[FuelType]string{
	FuelFlex:     "Flex",
	FuelGasolina: "Gasolina",
	FuelEtanol:   "Etanol",
	FuelDiesel:   "Diesel",
	FuelEletrico: "Elétrico",
	FuelHibrido:  "Híbrido",
}

var vehicleStatusOrder = []VehicleStatus{StatusActive, StatusSold, StatusInactive}

var vehicleStatusLabels = map[VehicleStatus]string{
	StatusActive:   "Ativo",
	StatusSold:     "Vendido",
	StatusInactive: "Inativo",
}

// fipeVehicleTypeItems: `FipeVehicleType` é texto livre no banco (segmento da
// tabela FIPE); o catálogo é a lista curada que o admin oferece.
var fipeVehicleTypeItems = []catalog.Item{
	{Slug: "carros", Name: "Carro"},
	{Slug: "motos", Name: "Moto"},
	{Slug: "caminhoes", Name: "Caminhão"},
}

// maintenanceTypeItems: o tipo da manutenção é texto livre no banco, gravado
// com o próprio rótulo como valor (dados legados); o catálogo é a lista
// curada que o admin oferece.
var maintenanceTypeItems = []catalog.Item{
	{Slug: "Troca de óleo", Name: "Troca de óleo"},
	{Slug: "Revisão preventiva", Name: "Revisão preventiva"},
	{Slug: "Rodízio de pneus", Name: "Rodízio de pneus"},
	{Slug: "Alinhamento e balanceamento", Name: "Alinhamento e balanceamento"},
	{Slug: "Filtro de ar", Name: "Filtro de ar"},
	{Slug: "Filtro de cabine", Name: "Filtro de cabine"},
	{Slug: "Fluido de freio", Name: "Fluido de freio"},
	{Slug: "Pastilhas de freio", Name: "Pastilhas de freio"},
	{Slug: "Correia dentada", Name: "Correia dentada"},
	{Slug: "Velas de ignição", Name: "Velas de ignição"},
	{Slug: "Fluido de arrefecimento", Name: "Fluido de arrefecimento"},
	{Slug: "Outro", Name: "Outro"},
}

var maintenanceStatusOrder = []MaintenanceStatus{
	MaintenanceStatusOrcado, MaintenanceStatusAgendado, MaintenanceStatusEmAndamento,
	MaintenanceStatusRealizado, MaintenanceStatusCancelado,
}

var maintenanceStatusLabels = map[MaintenanceStatus]string{
	MaintenanceStatusOrcado:      "Orçado",
	MaintenanceStatusAgendado:    "Agendado",
	MaintenanceStatusEmAndamento: "Em andamento",
	MaintenanceStatusRealizado:   "Realizado",
	MaintenanceStatusCancelado:   "Cancelado",
}

var osItemTypeOrder = []OSItemType{OSItemProduct, OSItemService}

var osItemTypeLabels = map[OSItemType]string{
	OSItemProduct: "Peça / Produto",
	OSItemService: "Serviço",
}

var osItemCategoryOrder = []OSItemCategory{
	OSItemCategoryMotor, OSItemCategoryFreios, OSItemCategorySuspensao, OSItemCategoryTransmissao,
	OSItemCategoryArrefecimento, OSItemCategoryEletrico, OSItemCategoryPneus, OSItemCategoryArCondicionado,
	OSItemCategoryCarroceria, OSItemCategoryServico, OSItemCategoryOutros,
}

var osItemCategoryLabels = map[OSItemCategory]string{
	OSItemCategoryMotor:          "Motor",
	OSItemCategoryFreios:         "Freios",
	OSItemCategorySuspensao:      "Suspensão",
	OSItemCategoryTransmissao:    "Transmissão",
	OSItemCategoryArrefecimento:  "Arrefecimento",
	OSItemCategoryEletrico:       "Elétrico",
	OSItemCategoryPneus:          "Pneus",
	OSItemCategoryArCondicionado: "Ar-condicionado",
	OSItemCategoryCarroceria:     "Carroceria",
	OSItemCategoryServico:        "Serviço",
	OSItemCategoryOutros:         "Outros",
}

// paymentMethodItems: `PaymentMethod` da manutenção é texto livre no banco e
// NÃO é a lista do financeiro (`cartao_debito` aqui vs `debito` lá); unificar
// exige migrar os registros existentes. Até lá, esta é a lista curada da frota.
var paymentMethodItems = []catalog.Item{
	{Slug: "pix", Name: "Pix"},
	{Slug: "cartao_credito", Name: "Cartão de Crédito"},
	{Slug: "cartao_debito", Name: "Cartão de Débito"},
	{Slug: "dinheiro", Name: "Dinheiro"},
	{Slug: "boleto", Name: "Boleto"},
	{Slug: "transferencia", Name: "Transferência"},
}

var alertStatusOrder = []AlertStatus{AlertOverdue, AlertDueSoon, AlertOK, AlertUnknown}

var alertStatusLabels = map[AlertStatus]string{
	AlertOverdue: "Vencido",
	AlertDueSoon: "Próximo do vencimento",
	AlertOK:      "Em dia",
	AlertUnknown: "Sem histórico",
}

var scheduleStatusOrder = []ScheduleStatus{
	ScheduleStatusPending, ScheduleStatusDueSoon, ScheduleStatusOverdue, ScheduleStatusDone, ScheduleStatusCancelled,
}

var scheduleStatusLabels = map[ScheduleStatus]string{
	ScheduleStatusPending:   "Pendente",
	ScheduleStatusDueSoon:   "Próximo do vencimento",
	ScheduleStatusOverdue:   "Vencido",
	ScheduleStatusDone:      "Concluído",
	ScheduleStatusCancelled: "Cancelado",
}

// Catalogs devolve os catálogos do módulo de frota, na ordem de exibição.
func Catalogs() []catalog.Catalog {
	return []catalog.Catalog{
		{Key: "vehicle.fuel_type", Items: catalog.FromLabels(fuelTypeOrder, fuelTypeLabels)},
		{Key: "vehicle.vehicle_status", Items: catalog.FromLabels(vehicleStatusOrder, vehicleStatusLabels)},
		{Key: "vehicle.fipe_vehicle_type", Items: fipeVehicleTypeItems},
		{Key: "vehicle.maintenance_type", Items: maintenanceTypeItems},
		{Key: "vehicle.maintenance_status", Items: catalog.FromLabels(maintenanceStatusOrder, maintenanceStatusLabels)},
		{Key: "vehicle.os_item_type", Items: catalog.FromLabels(osItemTypeOrder, osItemTypeLabels)},
		{Key: "vehicle.os_item_category", Items: catalog.FromLabels(osItemCategoryOrder, osItemCategoryLabels)},
		{Key: "vehicle.payment_method", Items: paymentMethodItems},
		{Key: "vehicle.alert_status", Items: catalog.FromLabels(alertStatusOrder, alertStatusLabels)},
		{Key: "vehicle.schedule_alert_status", Items: catalog.FromLabels(scheduleStatusOrder, scheduleStatusLabels)},
	}
}
