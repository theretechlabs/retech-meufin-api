package finance

import "github.com/retechfin/retechfin-api/internal/domain/catalog"

// Catálogos do módulo financeiro publicados em GET /api/v1/catalogs
// (docs/catalogs.md). Novo valor num enum = nova constante + rótulo aqui; o
// teste do serviço de catálogo falha se faltar rótulo.

var incomeSourceKindOrder = []IncomeSourceKind{
	IncomeSourceCLT, IncomeSourcePJ, IncomeSourceFreelance, IncomeSourceRental,
	IncomeSourceInvestment, IncomeSourceBenefit, IncomeSourceOther,
}

var incomeSourceKindLabels = map[IncomeSourceKind]string{
	IncomeSourceCLT:        "CLT",
	IncomeSourcePJ:         "PJ",
	IncomeSourceFreelance:  "Projeto avulso",
	IncomeSourceRental:     "Aluguel",
	IncomeSourceInvestment: "Investimento",
	IncomeSourceBenefit:    "Benefício",
	IncomeSourceOther:      "Outro",
}

// incomeTypeItems é o tipo da receita (natureza do lançamento), gravado como
// texto livre em `Type` do lançamento de crédito. O catálogo é a lista curada
// que o admin oferece — curada (nunca livre) para a semântica ativa/passiva
// não degradar.
var incomeTypeItems = []catalog.Item{
	{Slug: "salario", Name: "Salário (CLT)"},
	{Slug: "pj_contrato", Name: "Remuneração PJ (contrato)"},
	{Slug: "pro_labore", Name: "Pró-labore"},
	{Slug: "dividendos", Name: "Dividendos"},
	{Slug: "rendimento", Name: "Rendimentos (juros/investimentos)"},
	{Slug: "aluguel", Name: "Aluguel"},
	{Slug: "freela", Name: "Freela"},
	{Slug: "ferias_13", Name: "13º/Férias"},
	{Slug: "beneficio", Name: "Benefício"},
	{Slug: "reembolso", Name: "Reembolso"},
	{Slug: "venda_bem", Name: "Venda de bem (veículo, imóvel)"},
	{Slug: "outro", Name: "Outro"},
}

var recurrenceOrder = []Recurrence{RecurrenceNone, RecurrenceWeekly, RecurrenceMonthly, RecurrenceYearly}

var recurrenceLabels = map[Recurrence]string{
	RecurrenceNone:    "Nenhuma",
	RecurrenceWeekly:  "Semanal",
	RecurrenceMonthly: "Mensal",
	RecurrenceYearly:  "Anual",
}

var statusOrder = []Status{StatusPrevista, StatusRealizada, StatusCancelada}

var statusLabels = map[Status]string{
	StatusPrevista:  "Prevista",
	StatusRealizada: "Realizada",
	StatusCancelada: "Cancelada",
}

var accountKindOrder = []AccountKind{AccountCorrente, AccountPoupanca, AccountCarteira, AccountDigital}

var accountKindLabels = map[AccountKind]string{
	AccountCorrente: "Conta corrente",
	AccountPoupanca: "Poupança",
	AccountCarteira: "Carteira / Dinheiro",
	AccountDigital:  "Conta digital",
}

// paymentMethodItems: `compensacao` fecha a lista porque só o sistema a grava
// (quitação/venda numa troca de bem); o admin precisa do rótulo para exibir
// lançamentos existentes, mas não a oferece na liquidação manual.
var paymentMethodItems = []catalog.Item{
	{Slug: string(PaymentPix), Name: "Pix"},
	{Slug: string(PaymentDebito), Name: "Débito"},
	{Slug: string(PaymentTransferencia), Name: "Transferência"},
	{Slug: string(PaymentBoleto), Name: "Boleto"},
	{Slug: string(PaymentDinheiro), Name: "Dinheiro"},
	{Slug: string(PaymentCartaoCredito), Name: "Cartão de crédito"},
	{Slug: string(PaymentCompensacao), Name: "Compensação (sem caixa)", Description: "Gravada pelo sistema na quitação ou venda de uma troca de bem; não é opção de liquidação manual"},
}

var attachmentTypeOrder = []AttachmentType{
	AttachmentBoleto, AttachmentPixQRCode, AttachmentNotaFiscal, AttachmentContrato, AttachmentFatura, AttachmentOutro,
}

var attachmentTypeLabels = map[AttachmentType]string{
	AttachmentBoleto:     "Boleto",
	AttachmentPixQRCode:  "QR Code Pix",
	AttachmentNotaFiscal: "Nota fiscal",
	AttachmentContrato:   "Contrato",
	AttachmentFatura:     "Fatura",
	AttachmentOutro:      "Outro",
}

var supplierCategoryOrder = []SupplierCategory{
	SupplierCategoryServicosPublicos, SupplierCategoryTelecom, SupplierCategoryStreaming,
	SupplierCategoryVarejo, SupplierCategoryFarmacia, SupplierCategorySaude, SupplierCategorySeguros,
	SupplierCategoryFinanceiro, SupplierCategoryEducacao, SupplierCategoryAlimentacao,
	SupplierCategoryTransporte, SupplierCategoryAcademia, SupplierCategoryMoradia,
	SupplierCategoryTecnologia, SupplierCategoryPet, SupplierCategoryJuridico, SupplierCategoryContabil,
	SupplierCategoryCondominio, SupplierCategoryVestuario, SupplierCategoryBeleza, SupplierCategoryViagem,
	SupplierCategoryEntretenimento,
	// Categorias automotivas (fornecedores das ordens de serviço da frota)
	SupplierCategoryAutoCenter, SupplierCategoryMecanica, SupplierCategoryPostoCombustivel, SupplierCategoryLojaPecas,
	SupplierCategoryOutros,
}

var supplierCategoryLabels = map[SupplierCategory]string{
	SupplierCategoryServicosPublicos: "Serviços Públicos",
	SupplierCategoryTelecom:          "Telecom",
	SupplierCategoryStreaming:        "Streaming",
	SupplierCategoryVarejo:           "Varejo / E-commerce",
	SupplierCategoryFarmacia:         "Farmácia",
	SupplierCategorySaude:            "Saúde",
	SupplierCategorySeguros:          "Seguros",
	SupplierCategoryFinanceiro:       "Financeiro",
	SupplierCategoryEducacao:         "Educação",
	SupplierCategoryAlimentacao:      "Alimentação / Delivery",
	SupplierCategoryTransporte:       "Transporte",
	SupplierCategoryAcademia:         "Academia / Bem-estar",
	SupplierCategoryMoradia:          "Moradia / Imóvel",
	SupplierCategoryTecnologia:       "Tecnologia / SaaS / TI",
	SupplierCategoryPet:              "Pet / Veterinário",
	SupplierCategoryJuridico:         "Jurídico / Cartório",
	SupplierCategoryContabil:         "Contabilidade / Fiscal",
	SupplierCategoryCondominio:       "Condomínio / IPTU",
	SupplierCategoryVestuario:        "Vestuário / Calçados",
	SupplierCategoryBeleza:           "Beleza / Estética",
	SupplierCategoryViagem:           "Viagem / Hospedagem",
	SupplierCategoryEntretenimento:   "Entretenimento / Eventos",
	SupplierCategoryAutoCenter:       "Auto center",
	SupplierCategoryMecanica:         "Mecânica / Oficina",
	SupplierCategoryPostoCombustivel: "Posto de combustível",
	SupplierCategoryLojaPecas:        "Loja de peças",
	SupplierCategoryOutros:           "Outros",
}

var supplierBillingOrder = []SupplierBillingType{
	SupplierBillingBoleto, SupplierBillingPix, SupplierBillingCartaoCredito, SupplierBillingDebitoAutomtico,
	SupplierBillingDebito, SupplierBillingTransferencia, SupplierBillingDescontoFolha,
}

var supplierBillingLabels = map[SupplierBillingType]string{
	SupplierBillingBoleto:          "Boleto",
	SupplierBillingPix:             "Pix",
	SupplierBillingCartaoCredito:   "Cartão de crédito",
	SupplierBillingDebitoAutomtico: "Débito automático",
	SupplierBillingDebito:          "Débito",
	SupplierBillingTransferencia:   "Transferência",
	SupplierBillingDescontoFolha:   "Desconto em folha",
}

var renegotiationKindOrder = []RenegotiationKind{KindRenegotiation, KindPayoff, KindAssetSwap}

var renegotiationKindLabels = map[RenegotiationKind]string{
	KindRenegotiation: "Renegociação",
	KindPayoff:        "Quitação antecipada",
	KindAssetSwap:     "Troca de bem",
}

// cardBankItems: `Bank` do cartão é texto livre no banco; este é o catálogo
// curado de emissores que o admin oferece. A bandeira é `finance.card_brand`.
var cardBankItems = []catalog.Item{
	{Slug: "nubank", Name: "Nubank"},
	{Slug: "inter", Name: "Inter"},
	{Slug: "santander", Name: "Santander"},
	{Slug: "itau", Name: "Itaú"},
	{Slug: "bradesco", Name: "Bradesco"},
	{Slug: "banco_do_brasil", Name: "Banco do Brasil"},
	{Slug: "caixa", Name: "Caixa"},
	{Slug: "c6", Name: "C6"},
	{Slug: "btg", Name: "BTG Pactual"},
	{Slug: "xp", Name: "XP"},
	{Slug: "mercado_pago", Name: "Mercado Pago"},
	{Slug: "picpay", Name: "PicPay"},
	{Slug: "neon", Name: "Neon"},
	{Slug: "pan", Name: "Banco Pan"},
	{Slug: "original", Name: "Original"},
	{Slug: "safra", Name: "Safra"},
	{Slug: "sicoob", Name: "Sicoob"},
	{Slug: "sicredi", Name: "Sicredi"},
	{Slug: "banrisul", Name: "Banrisul"},
	{Slug: "bmg", Name: "BMG"},
	{Slug: "bv", Name: "BV (Votorantim)"},
	{Slug: "daycoval", Name: "Daycoval"},
	{Slug: "pagbank", Name: "PagBank (PagSeguro)"},
	{Slug: "will_bank", Name: "Will Bank"},
	{Slug: "agibank", Name: "Agibank"},
	{Slug: "digio", Name: "Digio"},
	{Slug: "porto_bank", Name: "Porto Bank (Porto Seguro)"},
	{Slug: "credicard", Name: "Credicard (Itaú)"},
	// Emissoras de cartão de varejo (private label / co-branded)
	{Slug: "csf", Name: "Banco CSF (Carrefour/Atacadão)"},
	{Slug: "bradescard", Name: "Bradescard (C&A e outros)"},
	{Slug: "cetelem", Name: "Cetelem (BNP Paribas)"},
	{Slug: "midway", Name: "Midway (Riachuelo)"},
	{Slug: "realize", Name: "Realize (Renner)"},
	{Slug: "pefisa", Name: "Pefisa (Pernambucanas)"},
	{Slug: "luizacred", Name: "Luizacred (Magalu)"},
	{Slug: "outro", Name: "Outro"},
}

// bankAccountTypeItems: `BankAccountType` do fornecedor é texto livre no
// banco; o catálogo é a lista curada que o admin oferece.
var bankAccountTypeItems = []catalog.Item{
	{Slug: "corrente", Name: "Corrente"},
	{Slug: "poupanca", Name: "Poupança"},
}

var assetTypeOrder = []AssetType{AssetVehicle, AssetProperty}

var assetTypeLabels = map[AssetType]string{
	AssetVehicle:  "Veículo",
	AssetProperty: "Imóvel",
}

func discountReasonItems() []catalog.Item {
	items := make([]catalog.Item, 0, len(DiscountReasons))
	for _, r := range DiscountReasons {
		items = append(items, catalog.Item{Slug: r.Slug, Name: r.Name, Description: r.Description})
	}
	return items
}

func cancelReasonItems() []catalog.Item {
	items := make([]catalog.Item, 0, len(CancelReasons))
	for _, r := range CancelReasons {
		items = append(items, catalog.Item{Slug: r.Slug, Name: r.Name, Description: r.Description})
	}
	return items
}

func cardBrandItems() []catalog.Item {
	items := make([]catalog.Item, 0, len(CardBrands))
	for _, b := range CardBrands {
		items = append(items, catalog.Item{Slug: b.Slug, Name: b.Name})
	}
	return items
}

// Catalogs devolve os catálogos do módulo financeiro, na ordem de exibição.
func Catalogs() []catalog.Catalog {
	return []catalog.Catalog{
		{Key: "finance.income_source_kind", Items: catalog.FromLabels(incomeSourceKindOrder, incomeSourceKindLabels)},
		{Key: "finance.income_type", Items: incomeTypeItems},
		{Key: "finance.recurrence", Items: catalog.FromLabels(recurrenceOrder, recurrenceLabels)},
		{Key: "finance.entry_status", Items: catalog.FromLabels(statusOrder, statusLabels)},
		{Key: "finance.account_kind", Items: catalog.FromLabels(accountKindOrder, accountKindLabels)},
		{Key: "finance.payment_method", Items: paymentMethodItems},
		{Key: "finance.attachment_type", Items: catalog.FromLabels(attachmentTypeOrder, attachmentTypeLabels)},
		{Key: "finance.supplier_category", Items: catalog.FromLabels(supplierCategoryOrder, supplierCategoryLabels)},
		{Key: "finance.supplier_billing_type", Items: catalog.FromLabels(supplierBillingOrder, supplierBillingLabels)},
		{Key: "finance.renegotiation_kind", Items: catalog.FromLabels(renegotiationKindOrder, renegotiationKindLabels)},
		{Key: "finance.card_bank", Items: cardBankItems},
		{Key: "finance.bank_account_type", Items: bankAccountTypeItems},
		{Key: "finance.asset_type", Items: catalog.FromLabels(assetTypeOrder, assetTypeLabels)},
		{Key: "finance.discount_reason", Items: discountReasonItems()},
		{Key: "finance.cancel_reason", Items: cancelReasonItems()},
		{Key: "finance.card_brand", Items: cardBrandItems()},
	}
}
