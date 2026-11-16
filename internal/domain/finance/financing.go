package finance

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MaxFinancingTermMonths é o prazo máximo aceito num contrato (50 anos).
const MaxFinancingTermMonths = 600

// FinancingCategorySlug é a categoria de despesa das parcelas de contrato.
const FinancingCategorySlug = "financiamentos"

// FinancingContract descreve o financiamento de um bem (imóvel ou veículo).
//
// As parcelas são lançamentos comuns, geradas pelo parcelamento; o contrato
// guarda o que a série não sabe — os dados do banco, a posição de referência
// usada para montar as parcelas e o saldo devedor informado.
type FinancingContract struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	AssetType   AssetType
	AssetID     uuid.UUID
	// GroupID é o recurrence_group_id da série RAIZ. Depois de uma
	// renegociação a série vigente é outra, alcançada pela linhagem.
	GroupID uuid.UUID
	// SupplierID é o banco credor (fornecedor), o mesmo das parcelas.
	SupplierID     *uuid.UUID
	Description    string
	ContractNumber *string
	PurchaseDate   *time.Time
	FinancedCents  *int64
	// DownPaymentCents é a entrada paga na compra (informacional).
	DownPaymentCents *int64
	TermMonths       int
	// Taxas anuais em ppm da fração (4,5939% a.a. = 45939). Basis points
	// não comportam as quatro casas que os bancos publicam.
	NominalRatePPM               *int
	EffectiveRatePPM             *int
	EffectiveRelationshipRatePPM *int
	// InstallmentCents é o valor base vigente da prestação (o que sai da
	// conta: amortização + juros + seguro + taxas).
	InstallmentCents int64
	// ReferenceDate/ReferenceInstallment: vencimento e número da parcela
	// usada como referência no cadastro. FirstDueDate deriva das duas.
	ReferenceDate        time.Time
	ReferenceInstallment int
	FirstDueDate         time.Time
	// BalanceCents é o saldo devedor informado pelo banco em BalanceDate.
	// Não é a soma das parcelas em aberto: a soma carrega juros futuros.
	BalanceCents *int64
	BalanceDate  *time.Time
	// AutoDebit: a prestação é debitada em conta; as parcelas vencidas a
	// partir de AutoDebitSince são liquidadas sozinhas, na data do vencimento.
	AutoDebit          bool
	AutoDebitSince     *time.Time
	AutoDebitAccountID *uuid.UUID
	Notes              *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// Validate verifica as invariantes do contrato.
func (c *FinancingContract) Validate() error {
	c.Description = strings.TrimSpace(c.Description)
	if c.WorkspaceID == uuid.Nil {
		return &ValidationError{Msg: "workspace_id é obrigatório"}
	}
	if !ValidAssetType(c.AssetType) || c.AssetID == uuid.Nil {
		return &ValidationError{Msg: "informe o bem financiado"}
	}
	if c.Description == "" {
		return &ValidationError{Msg: "informe a descrição do financiamento"}
	}
	if c.TermMonths < 2 || c.TermMonths > MaxFinancingTermMonths {
		return &ValidationError{Msg: "prazo contratado deve ficar entre 2 e 600 meses"}
	}
	if c.InstallmentCents <= 0 {
		return &ValidationError{Msg: "valor da prestação deve ser maior que zero"}
	}
	if c.ReferenceDate.IsZero() {
		return &ValidationError{Msg: "informe o vencimento da parcela atual"}
	}
	if c.ReferenceInstallment < 1 || c.ReferenceInstallment > c.TermMonths {
		return &ValidationError{Msg: "número da parcela atual deve ficar entre 1 e o prazo contratado"}
	}
	for _, v := range []*int64{c.FinancedCents, c.DownPaymentCents, c.BalanceCents} {
		if v != nil && *v < 0 {
			return &ValidationError{Msg: "valores do contrato não podem ser negativos"}
		}
	}
	for _, v := range []*int{c.NominalRatePPM, c.EffectiveRatePPM, c.EffectiveRelationshipRatePPM} {
		if v != nil && (*v < 0 || *v > 1_000_000) {
			return &ValidationError{Msg: "taxa de juros inválida"}
		}
	}
	if (c.BalanceCents == nil) != (c.BalanceDate == nil) {
		return &ValidationError{Msg: "saldo devedor e a data do saldo devem vir juntos"}
	}
	if c.AutoDebit && c.AutoDebitSince == nil {
		return &ValidationError{Msg: "débito automático exige a data de início"}
	}
	return nil
}

// RemainingMonths é o prazo restante na referência, contando a parcela atual
// (é como os bancos mostram: parcela 152 de 360 → 209 meses).
func (c *FinancingContract) RemainingMonths() int {
	return c.TermMonths - c.ReferenceInstallment + 1
}

// FinancingDueDate devolve o vencimento da parcela `number` de um contrato
// cuja parcela `refNumber` vence em `refDate`. O dia é o da referência,
// limitado ao último dia do mês (31 → 28/fev) sem contaminar os meses
// seguintes.
func FinancingDueDate(refDate time.Time, refNumber, number int) time.Time {
	loc := refDate.Location()
	if loc == nil {
		loc = time.UTC
	}
	months := refDate.Year()*12 + int(refDate.Month()) - 1 + (number - refNumber)
	return dateSameDayClamped(months/12, time.Month(months%12+1), refDate.Day(), loc)
}

// FinancingInstallments monta a série completa do contrato: as parcelas
// anteriores à de referência e as seguintes, todas pelo valor base. Com
// settlePast, as anteriores nascem pagas na própria data de vencimento.
func FinancingInstallments(refDate time.Time, refNumber, term int, amountCents int64, settlePast bool) []InstallmentSpec {
	specs := make([]InstallmentSpec, 0, term)
	for n := 1; n <= term; n++ {
		due := FinancingDueDate(refDate, refNumber, n)
		sp := InstallmentSpec{DueDate: due, AmountCents: amountCents}
		if settlePast && n < refNumber {
			paidAt := due
			sp.PaidAt = &paidAt
		}
		specs = append(specs, sp)
	}
	return specs
}

// FinancingContractRepository persiste contratos de financiamento.
type FinancingContractRepository interface {
	Create(ctx context.Context, c *FinancingContract) error
	Update(ctx context.Context, c *FinancingContract) error
	GetByID(ctx context.Context, workspaceID, id uuid.UUID) (*FinancingContract, error)
	ListByAsset(ctx context.Context, workspaceID uuid.UUID, assetType AssetType, assetID uuid.UUID) ([]FinancingContract, error)
	// ListAutoDebit devolve os contratos com débito automático de TODOS os
	// workspaces — insumo da rotina diária de liquidação.
	ListAutoDebit(ctx context.Context) ([]FinancingContract, error)
	// DeleteByGroup remove o contrato da série raiz indicada (a série foi
	// excluída). Não é erro não existir contrato.
	DeleteByGroup(ctx context.Context, workspaceID, groupID uuid.UUID) error
}
