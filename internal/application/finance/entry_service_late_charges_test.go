package finance

import (
	"context"
	"testing"
	"time"

	dom "github.com/retechfin/retechfin-api/internal/domain/finance"
)

func TestConfirmWithLateChargesAddsToPaid(t *testing.T) {
	repo := newFakeEntryRepo()
	e := seedEntry(repo, dom.StatusPrevista) // 10.000, vence 2026-07-10
	svc := NewFinancialEntryService(repo, fakeCategoryRepo{})

	fee, interest := int64(200), int64(50)
	paidAt := time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC)
	got, err := svc.Confirm(context.Background(), ConfirmEntryInput{
		WorkspaceID: e.WorkspaceID, ID: e.ID, PaidAt: &paidAt,
		LateFeeCents: &fee, LateInterestCents: &interest,
	})
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if got.PaidAmountCents == nil || *got.PaidAmountCents != 10_250 {
		t.Fatalf("pago default = valor + encargos (10250), veio %v", got.PaidAmountCents)
	}
	if got.LateFeeCents == nil || *got.LateFeeCents != 200 || got.LateInterestCents == nil || *got.LateInterestCents != 50 {
		t.Fatalf("encargos devem ficar gravados, veio %v/%v", got.LateFeeCents, got.LateInterestCents)
	}
	if got.AmountCents != 10_000 {
		t.Fatalf("valor original preservado, veio %d", got.AmountCents)
	}
	if findResidual(repo, e.ID) != nil {
		t.Fatal("pagamento integral com encargos não gera residual")
	}
}

func TestConfirmLateChargesWithExplicitPaidMatchingTotal(t *testing.T) {
	repo := newFakeEntryRepo()
	e := seedEntry(repo, dom.StatusPrevista)
	svc := NewFinancialEntryService(repo, fakeCategoryRepo{})

	fee := int64(300)
	paid := int64(10_300)
	got, err := svc.Confirm(context.Background(), ConfirmEntryInput{
		WorkspaceID: e.WorkspaceID, ID: e.ID, LateFeeCents: &fee, PaidAmountCents: &paid,
	})
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if *got.PaidAmountCents != 10_300 || findResidual(repo, e.ID) != nil {
		t.Fatalf("pago explícito igual ao devido com encargo: sem residual, veio %d", *got.PaidAmountCents)
	}
}

func TestConfirmPaidAboveDueStillRejectedWithoutCharges(t *testing.T) {
	repo := newFakeEntryRepo()
	e := seedEntry(repo, dom.StatusPrevista)
	svc := NewFinancialEntryService(repo, fakeCategoryRepo{})

	paid := int64(10_500)
	if _, err := svc.Confirm(context.Background(), ConfirmEntryInput{
		WorkspaceID: e.WorkspaceID, ID: e.ID, PaidAmountCents: &paid,
	}); err == nil {
		t.Fatal("pago acima do devido sem informar encargos deve falhar")
	}
}

func TestConfirmLateChargesRejectNegativeAndWaiver(t *testing.T) {
	repo := newFakeEntryRepo()
	e := seedEntry(repo, dom.StatusPrevista)
	svc := NewFinancialEntryService(repo, fakeCategoryRepo{})

	neg := int64(-1)
	if _, err := svc.Confirm(context.Background(), ConfirmEntryInput{
		WorkspaceID: e.WorkspaceID, ID: e.ID, LateFeeCents: &neg,
	}); err == nil {
		t.Fatal("multa negativa deve falhar")
	}
	// Isenção total + juros é contradição.
	discount := int64(10_000)
	reason := "cortesia"
	fee := int64(100)
	if _, err := svc.Confirm(context.Background(), ConfirmEntryInput{
		WorkspaceID: e.WorkspaceID, ID: e.ID, DiscountCents: &discount, DiscountReason: &reason, LateFeeCents: &fee,
	}); err == nil {
		t.Fatal("isenção total com encargo deve falhar")
	}
}

func TestConfirmZeroChargesStoredAsNil(t *testing.T) {
	repo := newFakeEntryRepo()
	e := seedEntry(repo, dom.StatusPrevista)
	svc := NewFinancialEntryService(repo, fakeCategoryRepo{})

	zero := int64(0)
	got, err := svc.Confirm(context.Background(), ConfirmEntryInput{
		WorkspaceID: e.WorkspaceID, ID: e.ID, LateFeeCents: &zero, LateInterestCents: &zero,
	})
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if got.LateFeeCents != nil || got.LateInterestCents != nil {
		t.Fatalf("zero não é encargo: deve gravar nil, veio %v/%v", got.LateFeeCents, got.LateInterestCents)
	}
	if *got.PaidAmountCents != 10_000 {
		t.Fatalf("pago = valor, veio %d", *got.PaidAmountCents)
	}
}

func TestSettleWithLateChargesDefaultsPaid(t *testing.T) {
	repo := newFakeEntryRepo()
	e := seedEntry(repo, dom.StatusPrevista)
	svc := NewFinancialEntryService(repo, fakeCategoryRepo{})

	fee, interest := int64(200), int64(100)
	discount := int64(500)
	reason := "cupom"
	got, err := svc.Settle(context.Background(), SettleEntryInput{
		WorkspaceID: e.WorkspaceID, ID: e.ID, PaymentMethod: "boleto",
		DiscountCents: &discount, DiscountReason: &reason,
		LateFeeCents: &fee, LateInterestCents: &interest,
	})
	if err != nil {
		t.Fatalf("Settle: %v", err)
	}
	// 10.000 − 500 + 200 + 100 = 9.800
	if *got.PaidAmountCents != 9_800 {
		t.Fatalf("pago default = valor − desconto + encargos (9800), veio %d", *got.PaidAmountCents)
	}
	if *got.LateFeeCents != 200 || *got.LateInterestCents != 100 {
		t.Fatalf("encargos gravados, veio %v/%v", got.LateFeeCents, got.LateInterestCents)
	}
}

func TestReopenClearsLateCharges(t *testing.T) {
	repo := newFakeEntryRepo()
	e := seedEntry(repo, dom.StatusPrevista)
	svc := NewFinancialEntryService(repo, fakeCategoryRepo{})

	fee := int64(200)
	if _, err := svc.Confirm(context.Background(), ConfirmEntryInput{
		WorkspaceID: e.WorkspaceID, ID: e.ID, LateFeeCents: &fee,
	}); err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	got, err := svc.Reopen(context.Background(), e.WorkspaceID, e.ID)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	if got.LateFeeCents != nil || got.LateInterestCents != nil {
		t.Fatalf("reopen deve limpar encargos, veio %v/%v", got.LateFeeCents, got.LateInterestCents)
	}
}

func TestLateChargeQuoteUsesEntryTermsAndDefaultsToToday(t *testing.T) {
	repo := newFakeEntryRepo()
	e := seedEntry(repo, dom.StatusPrevista) // vence 2026-07-10
	svc := NewFinancialEntryService(repo, fakeCategoryRepo{})

	at := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC) // 10 dias
	q, err := svc.LateChargeQuote(context.Background(), e.WorkspaceID, e.ID, &at)
	if err != nil {
		t.Fatalf("LateChargeQuote: %v", err)
	}
	if q.Source != dom.LateTermsDefault || q.DaysLate != 10 {
		t.Fatalf("quer padrão, 10 dias; veio %+v", q)
	}
	// 10.000: multa 2% = 200; juros 1% × 10/30 = 33.
	if q.LateFeeCents != 200 || q.LateInterestCents != 33 || q.TotalCents != 233 {
		t.Fatalf("quer 200/33/233, veio %+v", q)
	}

	// Termos próprios do lançamento prevalecem.
	feeBP, intBP := 500, 0
	e.LateFeeBP, e.LateInterestMonthlyBP = &feeBP, &intBP
	repo.entries[e.ID] = e
	q, err = svc.LateChargeQuote(context.Background(), e.WorkspaceID, e.ID, &at)
	if err != nil {
		t.Fatalf("LateChargeQuote: %v", err)
	}
	if q.Source != dom.LateTermsEntry || q.LateFeeCents != 500 || q.LateInterestCents != 0 {
		t.Fatalf("quer termos do lançamento 500/0, veio %+v", q)
	}

	// Sem paid_at: hoje (não dá para fixar o valor, mas não pode falhar).
	if _, err := svc.LateChargeQuote(context.Background(), e.WorkspaceID, e.ID, nil); err != nil {
		t.Fatalf("quote sem data: %v", err)
	}
}

func TestCreateInheritsLateTermsAcrossSeries(t *testing.T) {
	repo := newFakeEntryRepo()
	svc := NewFinancialEntryService(repo, fakeCategoryRepo{})
	feeBP := 300
	total := 3
	ws := seedEntry(repo, dom.StatusPrevista).WorkspaceID
	cat := "moradia"
	created, err := svc.Create(context.Background(), CreateEntryInput{
		WorkspaceID: ws, Kind: "debit", AmountCents: 30_000,
		DueDate: time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC), Type: &cat,
		Description: "aluguel", InstallmentsTotal: &total, LateFeeBP: &feeBP,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if len(created) != 3 {
		t.Fatalf("quer 3 parcelas, veio %d", len(created))
	}
	for _, c := range created {
		if c.LateFeeBP == nil || *c.LateFeeBP != 300 {
			t.Fatalf("parcela sem multa herdada: %+v", c.LateFeeBP)
		}
		// Só multa informada: juros gravados como zero, não como padrão.
		if c.LateInterestMonthlyBP == nil || *c.LateInterestMonthlyBP != 0 {
			t.Fatalf("juros devem ser zero explícito, veio %v", c.LateInterestMonthlyBP)
		}
		if c.LateFeeCents != nil || c.LateInterestCents != nil {
			t.Fatal("parcela nova não carrega encargo pago")
		}
	}
}
