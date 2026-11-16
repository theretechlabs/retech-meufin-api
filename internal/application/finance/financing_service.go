package finance

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"time"

	"github.com/google/uuid"
	dom "github.com/retechfin/retechfin-api/internal/domain/finance"
)

// brt é o fuso de negócio (Brasil não tem horário de verão desde 2019). O
// débito automático liquida no dia do vencimento daqui, não no de UTC.
var brt = time.FixedZone("BRT", -3*60*60)

// businessToday devolve a data de hoje no Brasil como meia-noite UTC — a
// mesma representação dos vencimentos (coluna DATE).
func businessToday() time.Time {
	n := time.Now().In(brt)
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
}

// FinancingService mantém contratos de financiamento de bens.
//
// O contrato não reimplementa nada do financeiro: as parcelas nascem pelo
// parcelamento (FinancialEntryService.Create), o reajuste é a edição em
// série, a baixa automática é a liquidação comum e a posição da dívida vem
// da linhagem (RenegotiationService). Aqui ficam só os dados do contrato e
// a tradução deles para essas operações.
type FinancingService struct {
	contracts  dom.FinancingContractRepository
	entries    *FinancialEntryService
	debts      *RenegotiationService
	categories *ExpenseCategoryService
	now        func() time.Time
}

func NewFinancingService(contracts dom.FinancingContractRepository, entries *FinancialEntryService, debts *RenegotiationService, categories *ExpenseCategoryService) *FinancingService {
	s := &FinancingService{contracts: contracts, entries: entries, debts: debts, categories: categories, now: businessToday}
	// Série excluída (por aqui ou pela tela de despesas) leva o contrato junto.
	entries.OnSeriesDeleted(contracts.DeleteByGroup)
	return s
}

// FinancingTerms são os dados do contrato que podem ser corrigidos depois.
type FinancingTerms struct {
	ContractNumber               *string
	PurchaseDate                 *time.Time
	FinancedCents                *int64
	DownPaymentCents             *int64
	NominalRatePPM               *int
	EffectiveRatePPM             *int
	EffectiveRelationshipRatePPM *int
	BalanceCents                 *int64
	BalanceDate                  *time.Time
	AutoDebit                    bool
	AutoDebitAccountID           *uuid.UUID
	Notes                        *string
}

type CreateFinancingInput struct {
	WorkspaceID uuid.UUID
	AssetType   dom.AssetType
	AssetID     uuid.UUID
	SupplierID  *uuid.UUID
	Description string
	TermMonths  int
	// InstallmentCents: valor base da prestação, usado em todas as parcelas.
	InstallmentCents int64
	// ReferenceDate/ReferenceInstallment: vencimento e número da parcela
	// atual. As anteriores e as seguintes são montadas a partir dela.
	ReferenceDate        time.Time
	ReferenceInstallment int
	// SettlePast: parcelas anteriores à de referência nascem pagas, pelo
	// valor base, na própria data de vencimento.
	SettlePast bool
	FinancingTerms
}

type CreateFinancingResult struct {
	Contract *dom.FinancingContract
	// Created/Settled: parcelas geradas e quantas nasceram pagas.
	Created int
	Settled int
}

// Create registra o contrato e gera a série inteira de parcelas.
func (s *FinancingService) Create(ctx context.Context, in CreateFinancingInput) (*CreateFinancingResult, error) {
	now := time.Now().UTC()
	c := &dom.FinancingContract{
		ID:                   uuid.New(),
		WorkspaceID:          in.WorkspaceID,
		AssetType:            in.AssetType,
		AssetID:              in.AssetID,
		SupplierID:           in.SupplierID,
		Description:          in.Description,
		TermMonths:           in.TermMonths,
		InstallmentCents:     in.InstallmentCents,
		ReferenceDate:        in.ReferenceDate,
		ReferenceInstallment: in.ReferenceInstallment,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	applyTerms(c, in.FinancingTerms)
	if c.AutoDebit {
		// A parcela de referência é a "atual": o banco debita dela em diante.
		since := in.ReferenceDate
		c.AutoDebitSince = &since
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	c.FirstDueDate = dom.FinancingDueDate(c.ReferenceDate, c.ReferenceInstallment, 1)

	// As categorias padrão são semeadas na primeira listagem; quem chega
	// pelo cadastro do bem pode nunca ter aberto o financeiro.
	if _, err := s.categories.List(ctx, c.WorkspaceID); err != nil {
		return nil, err
	}
	category := dom.FinancingCategorySlug
	assetType := c.AssetType
	created, err := s.entries.Create(ctx, CreateEntryInput{
		WorkspaceID:  c.WorkspaceID,
		Kind:         string(dom.KindDebit),
		Type:         s.entries.normalizeCategory(ctx, c.WorkspaceID, &category),
		Description:  c.Description,
		SupplierID:   c.SupplierID,
		AssetType:    &assetType,
		AssetID:      &c.AssetID,
		Installments: dom.FinancingInstallments(c.ReferenceDate, c.ReferenceInstallment, c.TermMonths, c.InstallmentCents, in.SettlePast),
	})
	if err != nil {
		return nil, err
	}
	c.GroupID = *created[0].RecurrenceGroupID
	if err := s.contracts.Create(ctx, c); err != nil {
		// Sem o contrato a série ficaria órfã no bem: desfaz as parcelas.
		if _, derr := s.entries.Delete(ctx, c.WorkspaceID, created[0].ID, DeleteScopeAll); derr != nil {
			slog.Error("financiamento: falha ao desfazer parcelas",
				slog.String("group_id", c.GroupID.String()), slog.String("error", derr.Error()))
		}
		return nil, err
	}
	res := &CreateFinancingResult{Contract: c, Created: len(created)}
	for i := range created {
		if created[i].Status == dom.StatusRealizada {
			res.Settled++
		}
	}
	if c.AutoDebit {
		n, err := s.settleDue(ctx, c, s.now())
		if err != nil {
			return nil, err
		}
		res.Settled += n
	}
	return res, nil
}

func applyTerms(c *dom.FinancingContract, t FinancingTerms) {
	c.ContractNumber = t.ContractNumber
	c.PurchaseDate = t.PurchaseDate
	c.FinancedCents = t.FinancedCents
	c.DownPaymentCents = t.DownPaymentCents
	c.NominalRatePPM = t.NominalRatePPM
	c.EffectiveRatePPM = t.EffectiveRatePPM
	c.EffectiveRelationshipRatePPM = t.EffectiveRelationshipRatePPM
	c.BalanceCents = t.BalanceCents
	c.BalanceDate = t.BalanceDate
	c.AutoDebit = t.AutoDebit
	c.AutoDebitAccountID = t.AutoDebitAccountID
	c.Notes = t.Notes
}

// Update corrige os dados do contrato. Prazo, prestação e parcela de
// referência definem a série já gerada e não mudam por aqui: valor da
// prestação muda por Readjust; o resto, excluindo e relançando o contrato.
func (s *FinancingService) Update(ctx context.Context, workspaceID, id uuid.UUID, t FinancingTerms) (*dom.FinancingContract, error) {
	c, err := s.contracts.GetByID(ctx, workspaceID, id)
	if err != nil {
		return nil, err
	}
	wasAuto, since := c.AutoDebit, c.AutoDebitSince
	applyTerms(c, t)
	switch {
	case !c.AutoDebit:
		c.AutoDebitSince = nil
		c.AutoDebitAccountID = nil
	case wasAuto:
		c.AutoDebitSince = since
	default:
		// Ligado agora: vale de hoje em diante. Parcelas vencidas antes
		// ficaram em aberto por decisão do usuário e não são tocadas.
		today := s.now()
		c.AutoDebitSince = &today
	}
	c.UpdatedAt = time.Now().UTC()
	if err := c.Validate(); err != nil {
		return nil, err
	}
	if err := s.contracts.Update(ctx, c); err != nil {
		return nil, err
	}
	if c.AutoDebit {
		if _, err := s.settleDue(ctx, c, s.now()); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// openInstallments devolve as parcelas em aberto da série VIGENTE do
// contrato (depois de uma renegociação ela não é mais a raiz), por
// vencimento. Residuais de pagamento parcial ficam de fora: não são parcelas.
func (s *FinancingService) openInstallments(ctx context.Context, c *dom.FinancingContract) ([]dom.FinancialEntry, error) {
	lineage, err := s.debts.Lineage(ctx, c.WorkspaceID, c.GroupID)
	if err != nil {
		return nil, err
	}
	if len(lineage.Stages) == 0 {
		return nil, nil
	}
	stage := lineage.Stages[len(lineage.Stages)-1]
	var open []dom.FinancialEntry
	for _, e := range stage.Entries {
		if e.Status == dom.StatusPrevista && e.InstallmentNumber != nil && e.ResidualOfID == nil {
			open = append(open, e)
		}
	}
	sort.SliceStable(open, func(i, j int) bool { return open[i].DueDate.Before(open[j].DueDate) })
	return open, nil
}

// Readjust aplica o novo valor da prestação às parcelas em aberto com
// vencimento a partir de `from` (correção pelo índice do contrato). Devolve
// quantas parcelas mudaram.
func (s *FinancingService) Readjust(ctx context.Context, workspaceID, id uuid.UUID, amountCents int64, from time.Time) (*dom.FinancingContract, int, error) {
	if amountCents <= 0 {
		return nil, 0, &dom.ValidationError{Msg: "valor da prestação deve ser maior que zero"}
	}
	c, err := s.contracts.GetByID(ctx, workspaceID, id)
	if err != nil {
		return nil, 0, err
	}
	open, err := s.openInstallments(ctx, c)
	if err != nil {
		return nil, 0, err
	}
	var targets []dom.FinancialEntry
	for _, e := range open {
		if !e.DueDate.Before(from) {
			targets = append(targets, e)
		}
	}
	if len(targets) == 0 {
		return nil, 0, &dom.ValidationError{Msg: "não há parcelas em aberto a partir dessa data"}
	}
	// A edição em série propaga o que mudou na âncora para as previstas
	// seguintes; a âncora é a primeira parcela que ainda não tem o valor novo.
	changed := 0
	for _, e := range targets {
		if e.AmountCents == amountCents {
			continue
		}
		_, stats, err := s.entries.Update(ctx, UpdateEntryInput{
			WorkspaceID:    e.WorkspaceID,
			ID:             e.ID,
			Kind:           string(e.Kind),
			AmountCents:    amountCents,
			DueDate:        e.DueDate,
			FamilyMemberID: e.FamilyMemberID,
			SourceID:       e.SourceID,
			Type:           e.Type,
			Description:    e.Description,
			Notes:          e.Notes,
			SupplierID:     e.SupplierID,
			ApplyToFuture:  true,
		})
		if err != nil {
			return nil, 0, err
		}
		changed = 1 + stats.Fields
		break
	}
	c.InstallmentCents = amountCents
	c.UpdatedAt = time.Now().UTC()
	if err := s.contracts.Update(ctx, c); err != nil {
		return nil, 0, err
	}
	return c, changed, nil
}

// Delete exclui o contrato e a série inteira de parcelas (inclusive as
// pagas) — é o desfazer de um cadastro errado. Dívida que já passou por
// renegociação, quitação ou troca tem história e não some por aqui.
func (s *FinancingService) Delete(ctx context.Context, workspaceID, id uuid.UUID) error {
	c, err := s.contracts.GetByID(ctx, workspaceID, id)
	if err != nil {
		return err
	}
	lineage, err := s.debts.Lineage(ctx, workspaceID, c.GroupID)
	if err != nil && !errors.Is(err, dom.ErrNotFound) {
		return err
	}
	if lineage != nil {
		if lineage.RenegotiationCnt > 0 || lineage.ClosedBy != nil || lineage.OriginEvent != nil {
			return &dom.ValidationError{Msg: "contrato renegociado, quitado ou vindo de troca não pode ser excluído"}
		}
		if anchor := firstInstallment(lineage); anchor != nil {
			if _, err := s.entries.Delete(ctx, workspaceID, anchor.ID, DeleteScopeAll); err != nil {
				return err
			}
		}
	}
	return s.contracts.DeleteByGroup(ctx, workspaceID, c.GroupID)
}

// firstInstallment devolve uma parcela da série (âncora da exclusão em
// série); residuais não servem, não carregam o grupo.
func firstInstallment(l *DebtLineage) *dom.FinancialEntry {
	for i := range l.Stages {
		for j := range l.Stages[i].Entries {
			e := &l.Stages[i].Entries[j]
			if e.InstallmentNumber != nil && e.ResidualOfID == nil && e.RecurrenceGroupID != nil {
				return e
			}
		}
	}
	return nil
}

// ListByAsset devolve os contratos do bem.
func (s *FinancingService) ListByAsset(ctx context.Context, workspaceID uuid.UUID, assetType dom.AssetType, assetID uuid.UUID) ([]dom.FinancingContract, error) {
	if !dom.ValidAssetType(assetType) {
		return nil, &dom.ValidationError{Msg: "asset_type inválido"}
	}
	return s.contracts.ListByAsset(ctx, workspaceID, assetType, assetID)
}

// settleDue liquida as parcelas em aberto vencidas entre o início do débito
// automático e hoje, cada uma na própria data de vencimento.
func (s *FinancingService) settleDue(ctx context.Context, c *dom.FinancingContract, today time.Time) (int, error) {
	if !c.AutoDebit || c.AutoDebitSince == nil {
		return 0, nil
	}
	open, err := s.openInstallments(ctx, c)
	if err != nil {
		if errors.Is(err, dom.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}
	settled := 0
	for _, e := range open {
		if e.DueDate.Before(*c.AutoDebitSince) || e.DueDate.After(today) {
			continue
		}
		paidAt := e.DueDate
		if _, err := s.entries.Settle(ctx, SettleEntryInput{
			WorkspaceID:   e.WorkspaceID,
			ID:            e.ID,
			PaidAt:        &paidAt,
			PaymentMethod: string(dom.PaymentDebito),
			AccountID:     c.AutoDebitAccountID,
		}); err != nil {
			return settled, err
		}
		settled++
	}
	return settled, nil
}

// SettleAutoDebits roda a baixa automática de todos os contratos com débito
// em conta. Roda no boot e diariamente (ticker no main). Falha num contrato
// não impede os demais.
func (s *FinancingService) SettleAutoDebits(ctx context.Context) (int, error) {
	contracts, err := s.contracts.ListAutoDebit(ctx)
	if err != nil {
		return 0, err
	}
	today := s.now()
	total := 0
	for i := range contracts {
		n, err := s.settleDue(ctx, &contracts[i], today)
		total += n
		if err != nil {
			slog.Error("débito automático: falha ao liquidar parcelas",
				slog.String("contract_id", contracts[i].ID.String()), slog.String("error", err.Error()))
		}
	}
	return total, nil
}
