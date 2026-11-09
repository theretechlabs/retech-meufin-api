package finance

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	dom "github.com/retechfin/retechfin-api/internal/domain/finance"
)

// seedInvoice cria fatura 'cartao' prevista com dois itens (um cancelado) e
// uma despesa avulsa fora da fatura.
func seedInvoice(repo *fakeEntryRepo) (invoice, item, cancelled *dom.FinancialEntry) {
	ws := uuid.New()
	due := time.Date(2026, 11, 13, 0, 0, 0, 0, time.UTC)
	cartao := dom.CartaoCategorySlug
	invoice = &dom.FinancialEntry{
		ID: uuid.New(), WorkspaceID: ws, Kind: dom.KindDebit, Status: dom.StatusPrevista,
		AmountCents: 1021457, DueDate: due, Type: &cartao, Description: "sicredi_123",
		Recurrence: dom.RecurrenceNone,
	}
	cat := "alimentacao"
	item = &dom.FinancialEntry{
		ID: uuid.New(), WorkspaceID: ws, Kind: dom.KindDebit, Status: dom.StatusPrevista,
		AmountCents: 10000, DueDate: due, Type: &cat, Description: "mercado",
		Recurrence: dom.RecurrenceNone, ParentID: &invoice.ID,
	}
	cancelled = &dom.FinancialEntry{
		ID: uuid.New(), WorkspaceID: ws, Kind: dom.KindDebit, Status: dom.StatusCancelada,
		AmountCents: 500, DueDate: due, Type: &cat, Description: "estorno",
		Recurrence: dom.RecurrenceNone, ParentID: &invoice.ID,
	}
	repo.entries[invoice.ID] = invoice
	repo.entries[item.ID] = item
	repo.entries[cancelled.ID] = cancelled
	return
}

func invoiceUpdateInput(e *dom.FinancialEntry) UpdateEntryInput {
	return UpdateEntryInput{
		WorkspaceID: e.WorkspaceID, ID: e.ID, Kind: string(e.Kind),
		AmountCents: e.AmountCents, DueDate: e.DueDate, Type: e.Type, Description: e.Description,
	}
}

func TestUpdateInvoiceDueDateCascadesToItems(t *testing.T) {
	repo := newFakeEntryRepo()
	inv, item, cancelled := seedInvoice(repo)
	svc := NewFinancialEntryService(repo, fakeCategoryRepo{})

	newDue := inv.DueDate.AddDate(0, 0, -3)
	in := invoiceUpdateInput(inv)
	in.DueDate = newDue
	got, _, err := svc.Update(context.Background(), in)
	if err != nil {
		t.Fatalf("editar vencimento da fatura: %v", err)
	}
	if !got.DueDate.Equal(newDue) || !got.IsCardInvoice() {
		t.Fatalf("fatura devia vencer em %s e continuar 'cartao'; got due=%s type=%v", newDue, got.DueDate, got.Type)
	}
	if len(repo.dueCascadeCalls) != 1 || repo.dueCascadeCalls[0].ParentID != inv.ID || !repo.dueCascadeCalls[0].DueDate.Equal(newDue) {
		t.Fatalf("esperava 1 cascata de vencimento para a fatura, got %+v", repo.dueCascadeCalls)
	}
	if !repo.entries[item.ID].DueDate.Equal(newDue) {
		t.Errorf("item devia acompanhar a fatura: %s", repo.entries[item.ID].DueDate)
	}
	if repo.entries[cancelled.ID].DueDate.Equal(newDue) {
		t.Errorf("item cancelado não deve ser tocado")
	}
}

func TestUpdateInvoiceLabelsDoNotCascade(t *testing.T) {
	repo := newFakeEntryRepo()
	inv, _, _ := seedInvoice(repo)
	svc := NewFinancialEntryService(repo, fakeCategoryRepo{})

	in := invoiceUpdateInput(inv)
	in.Description = "Fatura Sicredi novembro"
	if _, _, err := svc.Update(context.Background(), in); err != nil {
		t.Fatalf("editar descrição da fatura: %v", err)
	}
	if len(repo.dueCascadeCalls) != 0 {
		t.Errorf("sem mudança de vencimento não há cascata: %+v", repo.dueCascadeCalls)
	}
}

func TestUpdateInvoiceKeepsCartaoCategory(t *testing.T) {
	repo := newFakeEntryRepo()
	inv, _, _ := seedInvoice(repo)
	svc := NewFinancialEntryService(repo, fakeCategoryRepo{})

	other := "moradia"
	in := invoiceUpdateInput(inv)
	in.Type = &other
	if _, _, err := svc.Update(context.Background(), in); err == nil {
		t.Error("fatura não pode virar outra categoria")
	}
	in = invoiceUpdateInput(inv)
	in.Type = nil
	if _, _, err := svc.Update(context.Background(), in); err == nil {
		t.Error("fatura não pode perder a categoria 'cartao'")
	}
	in = invoiceUpdateInput(inv)
	in.Kind = string(dom.KindCredit)
	if _, _, err := svc.Update(context.Background(), in); err == nil {
		t.Error("fatura não pode virar receita")
	}
}

func TestUpdateRegularEntryCannotBecomeInvoice(t *testing.T) {
	repo := newFakeEntryRepo()
	e := seedEntry(repo, dom.StatusPrevista)
	svc := NewFinancialEntryService(repo, fakeCategoryRepo{})

	cartao := dom.CartaoCategorySlug
	in := invoiceUpdateInput(e)
	in.Type = &cartao
	if _, _, err := svc.Update(context.Background(), in); err == nil {
		t.Error("'cartao' continua reservado para lançamento comum")
	}
}

func TestUpdateInvoiceItemDueDateIsDerived(t *testing.T) {
	repo := newFakeEntryRepo()
	_, item, _ := seedInvoice(repo)
	svc := NewFinancialEntryService(repo, fakeCategoryRepo{})

	// Mesmo vencimento: edição do item passa (admin reenvia o due_date atual).
	in := invoiceUpdateInput(item)
	in.Description = "Mercado Zaffari"
	if _, _, err := svc.Update(context.Background(), in); err != nil {
		t.Fatalf("editar item mantendo vencimento: %v", err)
	}
	// Vencimento diferente: rejeita, corrige-se na fatura.
	in = invoiceUpdateInput(item)
	in.DueDate = item.DueDate.AddDate(0, 0, 5)
	if _, _, err := svc.Update(context.Background(), in); err == nil {
		t.Error("item de fatura não tem vencimento próprio")
	}
}

func TestUpdatePaidInvoiceDueDateStillBlocked(t *testing.T) {
	repo := newFakeEntryRepo()
	inv, _, _ := seedInvoice(repo)
	paid := time.Now().UTC()
	inv.Status = dom.StatusRealizada
	inv.PaidAt = &paid
	svc := NewFinancialEntryService(repo, fakeCategoryRepo{})

	in := invoiceUpdateInput(inv)
	in.DueDate = inv.DueDate.AddDate(0, 0, 1)
	if _, _, err := svc.Update(context.Background(), in); err == nil {
		t.Error("fatura paga não muda de vencimento sem desfazer o pagamento")
	}
	if len(repo.dueCascadeCalls) != 0 {
		t.Errorf("rejeição não pode cascatear: %+v", repo.dueCascadeCalls)
	}
}
