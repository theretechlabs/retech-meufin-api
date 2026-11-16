package finance

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	dom "github.com/retechfin/retechfin-api/internal/domain/finance"
)

type fakeFinancingRepo struct {
	contracts  map[uuid.UUID]*dom.FinancingContract
	failCreate error
}

func newFakeFinancingRepo() *fakeFinancingRepo {
	return &fakeFinancingRepo{contracts: map[uuid.UUID]*dom.FinancingContract{}}
}

func (f *fakeFinancingRepo) Create(_ context.Context, c *dom.FinancingContract) error {
	if f.failCreate != nil {
		return f.failCreate
	}
	cp := *c
	f.contracts[c.ID] = &cp
	return nil
}

func (f *fakeFinancingRepo) Update(_ context.Context, c *dom.FinancingContract) error {
	if _, ok := f.contracts[c.ID]; !ok {
		return dom.ErrNotFound
	}
	cp := *c
	f.contracts[c.ID] = &cp
	return nil
}

func (f *fakeFinancingRepo) GetByID(_ context.Context, workspaceID, id uuid.UUID) (*dom.FinancingContract, error) {
	c, ok := f.contracts[id]
	if !ok || c.WorkspaceID != workspaceID {
		return nil, dom.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (f *fakeFinancingRepo) ListByAsset(_ context.Context, workspaceID uuid.UUID, _ dom.AssetType, assetID uuid.UUID) ([]dom.FinancingContract, error) {
	var out []dom.FinancingContract
	for _, c := range f.contracts {
		if c.WorkspaceID == workspaceID && c.AssetID == assetID {
			out = append(out, *c)
		}
	}
	return out, nil
}

func (f *fakeFinancingRepo) ListAutoDebit(context.Context) ([]dom.FinancingContract, error) {
	var out []dom.FinancingContract
	for _, c := range f.contracts {
		if c.AutoDebit {
			out = append(out, *c)
		}
	}
	return out, nil
}

func (f *fakeFinancingRepo) DeleteByGroup(_ context.Context, workspaceID, groupID uuid.UUID) error {
	for id, c := range f.contracts {
		if c.WorkspaceID == workspaceID && c.GroupID == groupID {
			delete(f.contracts, id)
		}
	}
	return nil
}

type financingFixture struct {
	entries   *fakeEntryRepo
	contracts *fakeFinancingRepo
	svc       *FinancingService
	entrySvc  *FinancialEntryService
	ws        uuid.UUID
	asset     uuid.UUID
}

func newFinancingFixture(today time.Time) *financingFixture {
	entries := newFakeEntryRepo()
	contracts := newFakeFinancingRepo()
	entrySvc := NewFinancialEntryService(entries, fakeCategoryRepo{})
	renegSvc := NewRenegotiationService(entries, newFakeRenegRepo(entries))
	svc := NewFinancingService(contracts, entrySvc, renegSvc, NewExpenseCategoryService(fakeCategoryRepo{}))
	svc.now = func() time.Time { return today }
	return &financingFixture{entries: entries, contracts: contracts, svc: svc, entrySvc: entrySvc, ws: uuid.New(), asset: uuid.New()}
}

func day(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

// caixaInput reproduz o contrato do app do banco: 360 meses, parcela 152
// vencendo em 23/11/2026, prestação de R$ 429,92.
func (fx *financingFixture) caixaInput() CreateFinancingInput {
	return CreateFinancingInput{
		WorkspaceID:          fx.ws,
		AssetType:            dom.AssetProperty,
		AssetID:              fx.asset,
		Description:          "Financiamento Casa",
		TermMonths:           360,
		InstallmentCents:     42_992,
		ReferenceDate:        day(2026, 11, 23),
		ReferenceInstallment: 152,
		SettlePast:           true,
	}
}

// series devolve as parcelas do contrato por número.
func (fx *financingFixture) series(groupID uuid.UUID) []*dom.FinancialEntry {
	var out []*dom.FinancialEntry
	for _, e := range fx.entries.entries {
		if e.RecurrenceGroupID != nil && *e.RecurrenceGroupID == groupID && e.InstallmentNumber != nil {
			out = append(out, e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return *out[i].InstallmentNumber < *out[j].InstallmentNumber })
	return out
}

func TestFinancingCreateGeraSerieInteiraELiquidaAnteriores(t *testing.T) {
	fx := newFinancingFixture(day(2026, 11, 16))
	res, err := fx.svc.Create(context.Background(), fx.caixaInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if res.Created != 360 || res.Settled != 151 {
		t.Fatalf("created=%d settled=%d, quero 360/151", res.Created, res.Settled)
	}
	c := res.Contract
	if !c.FirstDueDate.Equal(day(2014, 4, 23)) {
		t.Fatalf("primeira parcela em %s, quero 2014-04-23", c.FirstDueDate.Format("2006-01-02"))
	}
	if c.RemainingMonths() != 209 {
		t.Fatalf("prazo restante %d, quero 209", c.RemainingMonths())
	}
	series := fx.series(c.GroupID)
	if len(series) != 360 {
		t.Fatalf("série com %d parcelas", len(series))
	}
	for _, e := range series {
		n := *e.InstallmentNumber
		if *e.InstallmentTotal != 360 || e.AmountCents != 42_992 {
			t.Fatalf("parcela %d: total=%d valor=%d", n, *e.InstallmentTotal, e.AmountCents)
		}
		if e.AssetType == nil || *e.AssetType != dom.AssetProperty || e.AssetID == nil || *e.AssetID != fx.asset {
			t.Fatalf("parcela %d sem vínculo com o imóvel", n)
		}
		if e.Type == nil || *e.Type != dom.FinancingCategorySlug {
			t.Fatalf("parcela %d com categoria %v", n, e.Type)
		}
		if n < 152 {
			if e.Status != dom.StatusRealizada || e.PaidAt == nil || !e.PaidAt.Equal(e.DueDate) || *e.PaidAmountCents != 42_992 {
				t.Fatalf("parcela %d deveria nascer paga no vencimento", n)
			}
		} else if e.Status != dom.StatusPrevista || e.PaidAt != nil {
			t.Fatalf("parcela %d deveria estar em aberto", n)
		}
	}
	if !series[151].DueDate.Equal(day(2026, 11, 23)) || !series[359].DueDate.Equal(day(2044, 3, 23)) {
		t.Fatalf("vencimentos: 152=%s 360=%s", series[151].DueDate.Format("2006-01-02"), series[359].DueDate.Format("2006-01-02"))
	}
}

func TestFinancingCreateSemLiquidarAnteriores(t *testing.T) {
	fx := newFinancingFixture(day(2026, 11, 16))
	in := fx.caixaInput()
	in.SettlePast = false
	res, err := fx.svc.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if res.Settled != 0 {
		t.Fatalf("settled=%d, quero 0", res.Settled)
	}
}

func TestFinancingDueDateMantemDiaComClampDeFimDeMes(t *testing.T) {
	ref := day(2026, 1, 31)
	if got := dom.FinancingDueDate(ref, 10, 11); !got.Equal(day(2026, 2, 28)) {
		t.Fatalf("fevereiro: %s", got.Format("2006-01-02"))
	}
	if got := dom.FinancingDueDate(ref, 10, 12); !got.Equal(day(2026, 3, 31)) {
		t.Fatalf("março deveria voltar ao dia 31: %s", got.Format("2006-01-02"))
	}
	if got := dom.FinancingDueDate(ref, 10, 9); !got.Equal(day(2025, 12, 31)) {
		t.Fatalf("mês anterior cruza o ano: %s", got.Format("2006-01-02"))
	}
}

func TestFinancingCreateValida(t *testing.T) {
	fx := newFinancingFixture(day(2026, 11, 16))
	cases := map[string]func(*CreateFinancingInput){
		"parcela além do prazo": func(in *CreateFinancingInput) { in.ReferenceInstallment = 361 },
		"prazo acima do máximo": func(in *CreateFinancingInput) { in.TermMonths = 601 },
		"prestação zerada":      func(in *CreateFinancingInput) { in.InstallmentCents = 0 },
		"sem descrição":         func(in *CreateFinancingInput) { in.Description = "  " },
		"saldo sem data":        func(in *CreateFinancingInput) { v := int64(1); in.BalanceCents = &v },
	}
	for name, mutate := range cases {
		in := fx.caixaInput()
		mutate(&in)
		if _, err := fx.svc.Create(context.Background(), in); !errors.Is(err, dom.ErrValidation) {
			t.Errorf("%s: quero erro de validação, veio %v", name, err)
		}
	}
	if len(fx.entries.entries) != 0 {
		t.Fatalf("entrada inválida não pode gerar parcelas: %d", len(fx.entries.entries))
	}
}

func TestFinancingCreateDesfazParcelasQuandoContratoFalha(t *testing.T) {
	fx := newFinancingFixture(day(2026, 11, 16))
	fx.contracts.failCreate = errors.New("banco fora")
	if _, err := fx.svc.Create(context.Background(), fx.caixaInput()); err == nil {
		t.Fatal("quero erro")
	}
	if len(fx.entries.entries) != 0 {
		t.Fatalf("parcelas órfãs: %d", len(fx.entries.entries))
	}
}

func TestFinancingReadjustAplicaAsParcelasEmAbertoDaDataEmDiante(t *testing.T) {
	fx := newFinancingFixture(day(2026, 11, 16))
	res, _ := fx.svc.Create(context.Background(), fx.caixaInput())
	c, changed, err := fx.svc.Readjust(context.Background(), fx.ws, res.Contract.ID, 43_500, day(2027, 1, 1))
	if err != nil {
		t.Fatalf("Readjust: %v", err)
	}
	// Parcelas 154 (23/01/2027) a 360.
	if changed != 207 || c.InstallmentCents != 43_500 {
		t.Fatalf("changed=%d prestação=%d", changed, c.InstallmentCents)
	}
	for _, e := range fx.series(c.GroupID) {
		want := int64(42_992)
		if *e.InstallmentNumber >= 154 {
			want = 43_500
		}
		if e.AmountCents != want {
			t.Fatalf("parcela %d com valor %d, quero %d", *e.InstallmentNumber, e.AmountCents, want)
		}
	}
	// Repetir o mesmo reajuste não muda nada nem falha.
	if _, changed, err := fx.svc.Readjust(context.Background(), fx.ws, c.ID, 43_500, day(2027, 1, 1)); err != nil || changed != 0 {
		t.Fatalf("reajuste repetido: changed=%d err=%v", changed, err)
	}
	if _, _, err := fx.svc.Readjust(context.Background(), fx.ws, c.ID, 43_500, day(2050, 1, 1)); !errors.Is(err, dom.ErrValidation) {
		t.Fatalf("sem parcelas na data: quero validação, veio %v", err)
	}
}

func TestFinancingAutoDebitLiquidaNoVencimento(t *testing.T) {
	fx := newFinancingFixture(day(2026, 11, 16))
	in := fx.caixaInput()
	in.AutoDebit = true
	account := uuid.New()
	in.AutoDebitAccountID = &account
	res, err := fx.svc.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if res.Settled != 151 {
		t.Fatalf("antes do vencimento nada é debitado: settled=%d", res.Settled)
	}
	// Dois vencimentos depois: 23/11 e 23/12.
	fx.svc.now = func() time.Time { return day(2026, 12, 23) }
	n, err := fx.svc.SettleAutoDebits(context.Background())
	if err != nil || n != 2 {
		t.Fatalf("SettleAutoDebits: n=%d err=%v", n, err)
	}
	series := fx.series(res.Contract.GroupID)
	for _, idx := range []int{151, 152} {
		e := series[idx]
		if e.Status != dom.StatusRealizada || !e.PaidAt.Equal(e.DueDate) {
			t.Fatalf("parcela %d deveria estar paga no vencimento", idx+1)
		}
		if e.PaymentMethod == nil || *e.PaymentMethod != dom.PaymentDebito || e.PaymentAccountID == nil || *e.PaymentAccountID != account {
			t.Fatalf("parcela %d sem forma de pagamento/conta do débito", idx+1)
		}
	}
	if series[153].Status != dom.StatusPrevista {
		t.Fatal("parcela futura não pode ser debitada")
	}
	if n, _ := fx.svc.SettleAutoDebits(context.Background()); n != 0 {
		t.Fatalf("rodar de novo no mesmo dia não liquida nada: %d", n)
	}
}

func TestFinancingAutoDebitLigadoDepoisNaoTocaVencidasAntigas(t *testing.T) {
	fx := newFinancingFixture(day(2026, 11, 16))
	in := fx.caixaInput()
	in.SettlePast = false
	res, _ := fx.svc.Create(context.Background(), in)

	fx.svc.now = func() time.Time { return day(2026, 12, 1) }
	c, err := fx.svc.Update(context.Background(), fx.ws, res.Contract.ID, FinancingTerms{AutoDebit: true})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if c.AutoDebitSince == nil || !c.AutoDebitSince.Equal(day(2026, 12, 1)) {
		t.Fatalf("início do débito: %v", c.AutoDebitSince)
	}
	fx.svc.now = func() time.Time { return day(2026, 12, 23) }
	if n, _ := fx.svc.SettleAutoDebits(context.Background()); n != 1 {
		t.Fatalf("só a parcela de 23/12 é debitada: %d", n)
	}
	series := fx.series(c.GroupID)
	if series[151].Status != dom.StatusPrevista || series[152].Status != dom.StatusRealizada {
		t.Fatal("a vencida antes de ligar o débito fica em aberto; a seguinte é paga")
	}
	// Desligar limpa o início e a conta.
	c, err = fx.svc.Update(context.Background(), fx.ws, c.ID, FinancingTerms{})
	if err != nil || c.AutoDebit || c.AutoDebitSince != nil {
		t.Fatalf("desligar: %+v err=%v", c, err)
	}
}

func TestFinancingDeleteRemoveContratoESerie(t *testing.T) {
	fx := newFinancingFixture(day(2026, 11, 16))
	res, _ := fx.svc.Create(context.Background(), fx.caixaInput())
	if err := fx.svc.Delete(context.Background(), fx.ws, res.Contract.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if len(fx.contracts.contracts) != 0 || len(fx.series(res.Contract.GroupID)) != 0 {
		t.Fatal("contrato e parcelas deveriam ter sido removidos")
	}
}

func TestFinancingContratoSomeQuandoSerieEExcluidaPelasDespesas(t *testing.T) {
	fx := newFinancingFixture(day(2026, 11, 16))
	res, _ := fx.svc.Create(context.Background(), fx.caixaInput())
	anchor := fx.series(res.Contract.GroupID)[200]
	if _, err := fx.entrySvc.Delete(context.Background(), fx.ws, anchor.ID, DeleteScopeAll); err != nil {
		t.Fatalf("Delete série: %v", err)
	}
	if len(fx.contracts.contracts) != 0 {
		t.Fatal("contrato órfão depois de excluir a série")
	}
}

func TestFinancingDeleteRecusaContratoRenegociado(t *testing.T) {
	fx := newFinancingFixture(day(2026, 11, 16))
	res, _ := fx.svc.Create(context.Background(), fx.caixaInput())
	if _, err := fx.svc.debts.Renegotiate(context.Background(), RenegotiateInput{
		WorkspaceID:      fx.ws,
		GroupID:          res.Contract.GroupID,
		InstallmentCount: 100,
		InstallmentCents: 50_000,
		FirstDueDate:     day(2026, 12, 23),
	}); err != nil {
		t.Fatalf("Renegotiate: %v", err)
	}
	if err := fx.svc.Delete(context.Background(), fx.ws, res.Contract.ID); !errors.Is(err, dom.ErrValidation) {
		t.Fatalf("quero validação, veio %v", err)
	}
	// O reajuste alcança a série vigente (a do acordo), não a raiz.
	_, changed, err := fx.svc.Readjust(context.Background(), fx.ws, res.Contract.ID, 51_000, day(2026, 12, 1))
	if err != nil || changed != 100 {
		t.Fatalf("reajuste pós-renegociação: changed=%d err=%v", changed, err)
	}
}
