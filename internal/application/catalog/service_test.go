package catalog

import (
	"regexp"
	"testing"

	domfin "github.com/retechfin/retechfin-api/internal/domain/finance"
	domhealth "github.com/retechfin/retechfin-api/internal/domain/health"
	domwar "github.com/retechfin/retechfin-api/internal/domain/warranty"
)

// adminKeys espelha o `CatalogKey` do admin (src/lib/catalogs/api.ts). Chave
// nova lá sem catálogo aqui derruba este teste — é o contrato.
var adminKeys = []string{
	// finance
	"finance.income_source_kind",
	"finance.income_type",
	"finance.recurrence",
	"finance.entry_status",
	"finance.account_kind",
	"finance.payment_method",
	"finance.attachment_type",
	"finance.supplier_category",
	"finance.supplier_billing_type",
	"finance.renegotiation_kind",
	"finance.card_bank",
	"finance.bank_account_type",
	"finance.asset_type",
	"finance.discount_reason",
	"finance.cancel_reason",
	"finance.card_brand",
	// education
	"education.stage",
	"education.shift",
	"education.list_status",
	"education.item_category",
	// health
	"health.relationship",
	"health.gender",
	"health.lab_kind",
	"health.plan_type",
	"health.appointment_kind",
	"health.specialty",
	"health.appointment_status",
	"health.member_doc_type",
	"health.plan_doc_type",
	"health.document_type",
	"health.exam_source_type",
	"health.exam_result_status",
	"health.marker_category",
	// homesafety
	"homesafety.category",
	"homesafety.risk_type",
	"homesafety.priority",
	"homesafety.event_type",
	"homesafety.status",
	// patrimony
	"patrimony.property_type",
	"patrimony.property_doc_type",
	"patrimony.asset_type",
	"patrimony.tax_type",
	"patrimony.tax_status",
	// vehicle
	"vehicle.fuel_type",
	"vehicle.vehicle_status",
	"vehicle.fipe_vehicle_type",
	"vehicle.maintenance_type",
	"vehicle.maintenance_status",
	"vehicle.os_item_type",
	"vehicle.os_item_category",
	"vehicle.payment_method",
	"vehicle.alert_status",
	"vehicle.schedule_alert_status",
	// warranty
	"warranty.category",
	"warranty.status",
	"warranty.doc_type",
}

func TestNew_ServesEveryAdminKey(t *testing.T) {
	svc, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	resp := svc.All()
	for _, k := range adminKeys {
		items, ok := resp.Catalogs[k]
		if !ok {
			t.Errorf("chave %q do admin não é servida", k)
			continue
		}
		if len(items) == 0 {
			t.Errorf("chave %q sem itens", k)
		}
	}
	if len(resp.Catalogs) != len(adminKeys) {
		extra := make([]string, 0)
		known := make(map[string]bool, len(adminKeys))
		for _, k := range adminKeys {
			known[k] = true
		}
		for k := range resp.Catalogs {
			if !known[k] {
				extra = append(extra, k)
			}
		}
		t.Errorf("API serve %d chaves, admin conhece %d; chaves sem par no admin: %v", len(resp.Catalogs), len(adminKeys), extra)
	}
}

func TestRegistry_KeysUniqueAndWellFormed(t *testing.T) {
	keyRe := regexp.MustCompile(`^[a-z]+\.[a-z_]+$`)
	seen := make(map[string]bool)
	for _, c := range registry() {
		if !keyRe.MatchString(c.Key) {
			t.Errorf("chave %q fora do padrão <modulo>.<enum>", c.Key)
		}
		if seen[c.Key] {
			t.Errorf("chave %q duplicada", c.Key)
		}
		seen[c.Key] = true
		slugs := make(map[string]bool, len(c.Items))
		for _, it := range c.Items {
			if it.Slug == "" || it.Name == "" {
				t.Errorf("%s: item %+v com slug/nome vazio", c.Key, it)
			}
			if slugs[it.Slug] {
				t.Errorf("%s: slug %q duplicado", c.Key, it.Slug)
			}
			slugs[it.Slug] = true
		}
	}
}

// Todo slug publicado tem de passar no validador do enum correspondente — o
// catálogo é a lista que o admin envia de volta.
func TestCatalog_SlugsPassDomainValidators(t *testing.T) {
	svc := MustNew()
	cats := svc.All().Catalogs

	checks := []struct {
		key   string
		valid func(string) bool
	}{
		{"finance.payment_method", func(s string) bool { return domfin.ValidPaymentMethod(domfin.PaymentMethod(s)) }},
		{"finance.attachment_type", func(s string) bool { return domfin.ValidAttachmentType(domfin.AttachmentType(s)) }},
		{"finance.asset_type", func(s string) bool { return domfin.ValidAssetType(domfin.AssetType(s)) }},
		{"finance.discount_reason", domfin.ValidDiscountReason},
		{"finance.cancel_reason", domfin.ValidCancelReason},
		{"finance.card_brand", domfin.ValidCardBrand},
		{"health.document_type", func(s string) bool { return domhealth.ValidDocumentType(domhealth.DocumentType(s)) }},
		{"health.member_doc_type", func(s string) bool { return domhealth.ValidMemberDocType(domhealth.MemberDocType(s)) }},
		{"health.plan_doc_type", func(s string) bool { return domhealth.ValidPlanDocType(domhealth.PlanDocType(s)) }},
		{"warranty.category", func(s string) bool { return domwar.ValidCategory(domwar.Category(s)) }},
		{"warranty.doc_type", func(s string) bool { return domwar.ValidDocType(domwar.DocType(s)) }},
	}
	for _, ck := range checks {
		items, ok := cats[ck.key]
		if !ok {
			t.Errorf("%s: catálogo ausente", ck.key)
			continue
		}
		for _, it := range items {
			if !ck.valid(it.Slug) {
				t.Errorf("%s: slug %q não passa no validador do domínio", ck.key, it.Slug)
			}
		}
	}
}

// Os catálogos curados em Go (motivos, bandeiras) são a fonte: o catálogo
// unificado tem de ter exatamente os mesmos slugs, na mesma ordem.
func TestCatalog_ReusesCuratedFinanceLists(t *testing.T) {
	cats := MustNew().All().Catalogs

	if got, want := cats["finance.discount_reason"], domfin.DiscountReasons; len(got) != len(want) {
		t.Fatalf("discount_reason: %d itens, esperado %d", len(got), len(want))
	} else {
		for i := range want {
			if got[i].Slug != want[i].Slug || got[i].Name != want[i].Name || got[i].Description != want[i].Description {
				t.Errorf("discount_reason[%d]: %+v != %+v", i, got[i], want[i])
			}
		}
	}
	if got, want := cats["finance.cancel_reason"], domfin.CancelReasons; len(got) != len(want) {
		t.Fatalf("cancel_reason: %d itens, esperado %d", len(got), len(want))
	} else {
		for i := range want {
			if got[i].Slug != want[i].Slug || got[i].Name != want[i].Name || got[i].Description != want[i].Description {
				t.Errorf("cancel_reason[%d]: %+v != %+v", i, got[i], want[i])
			}
		}
	}
	if got, want := cats["finance.card_brand"], domfin.CardBrands; len(got) != len(want) {
		t.Fatalf("card_brand: %d itens, esperado %d", len(got), len(want))
	} else {
		for i := range want {
			if got[i].Slug != want[i].Slug || got[i].Name != want[i].Name {
				t.Errorf("card_brand[%d]: %+v != %+v", i, got[i], want[i])
			}
		}
	}
}

func TestVersion_DeterministicAndShaped(t *testing.T) {
	a := MustNew()
	b := MustNew()
	if a.Version() != b.Version() {
		t.Fatalf("versão não determinística: %q vs %q", a.Version(), b.Version())
	}
	if !regexp.MustCompile(`^[0-9a-f]{16}$`).MatchString(a.Version()) {
		t.Fatalf("versão %q não é 16 hex", a.Version())
	}
	if a.All().Version != a.Version() {
		t.Fatalf("Response.Version %q != Version() %q", a.All().Version, a.Version())
	}
}

// Slugs são o valor que o admin grava hoje, verbatim — nunca re-slugificados.
// Pinamos os casos onde a coluna é texto livre (um slug diferente quebraria o
// match com registros existentes em silêncio).
func TestCatalog_SlugsAreVerbatimStoredValues(t *testing.T) {
	cats := MustNew().All().Catalogs
	pins := map[string][]string{
		"vehicle.maintenance_type":    {"Troca de óleo", "Revisão preventiva", "Alinhamento e balanceamento", "Outro"},
		"vehicle.payment_method":      {"pix", "cartao_credito", "cartao_debito", "dinheiro", "boleto", "transferencia"},
		"vehicle.fipe_vehicle_type":   {"carros", "motos", "caminhoes"},
		"health.gender":               {"male", "female", "other"},
		"health.marker_category":      {"hematologia", "bioquimica", "lipidico", "hepatico", "renal", "eletrolitos", "hormonios", "vitaminas", "inflamacao", "outros"},
		"finance.card_bank":           {"nubank", "banco_do_brasil", "bv", "pagbank", "csf", "outro"},
		"finance.income_type":         {"salario", "pj_contrato", "pro_labore", "ferias_13", "venda_bem", "outro"},
		"finance.bank_account_type":   {"corrente", "poupanca"},
		"health.member_doc_type":      {"cpf", "carteira_trabalho", "certidao_nascimento", "cartao_sus", "outro"},
		"health.plan_doc_type":        {"contrato", "tabela_coparticipacao", "declaracao_ir", "outro"},
		"health.document_type":        {"exam_result", "exam_request", "image_report", "medical_report", "prescription", "other"},
		"patrimony.property_doc_type": {"escritura", "matricula", "iptu", "outros"},
		"warranty.doc_type":           {"nota_fiscal", "certificado", "garantia_estendida", "manual", "outros"},
	}
	for key, want := range pins {
		items, ok := cats[key]
		if !ok {
			t.Errorf("%s: catálogo ausente", key)
			continue
		}
		have := make(map[string]bool, len(items))
		for _, it := range items {
			have[it.Slug] = true
		}
		for _, slug := range want {
			if !have[slug] {
				t.Errorf("%s: slug %q (valor gravado pelo admin) ausente", key, slug)
			}
		}
	}
}
