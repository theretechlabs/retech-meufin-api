package persistence

import (
	"context"
	"time"

	"github.com/google/uuid"
	dom "github.com/retechfin/retechfin-api/internal/domain/finance"
	"gorm.io/gorm"
)

// FinancingContractModel mapeia finance_financing_contracts.
type FinancingContractModel struct {
	ID                           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	WorkspaceID                  uuid.UUID  `gorm:"type:uuid;not null"`
	AssetType                    string     `gorm:"column:asset_type;size:20;not null"`
	AssetID                      uuid.UUID  `gorm:"column:asset_id;type:uuid;not null"`
	GroupID                      uuid.UUID  `gorm:"column:group_id;type:uuid;not null"`
	SupplierID                   *uuid.UUID `gorm:"column:supplier_id;type:uuid"`
	Description                  string     `gorm:"size:255;not null"`
	ContractNumber               *string    `gorm:"column:contract_number;size:60"`
	PurchaseDate                 *time.Time `gorm:"column:purchase_date;type:date"`
	FinancedCents                *int64     `gorm:"column:financed_cents"`
	DownPaymentCents             *int64     `gorm:"column:down_payment_cents"`
	TermMonths                   int        `gorm:"column:term_months;not null"`
	NominalRatePPM               *int       `gorm:"column:nominal_rate_ppm"`
	EffectiveRatePPM             *int       `gorm:"column:effective_rate_ppm"`
	EffectiveRelationshipRatePPM *int       `gorm:"column:effective_relationship_rate_ppm"`
	InstallmentCents             int64      `gorm:"column:installment_cents;not null"`
	ReferenceDate                time.Time  `gorm:"column:reference_date;type:date;not null"`
	ReferenceInstallment         int        `gorm:"column:reference_installment;not null"`
	FirstDueDate                 time.Time  `gorm:"column:first_due_date;type:date;not null"`
	BalanceCents                 *int64     `gorm:"column:balance_cents"`
	BalanceDate                  *time.Time `gorm:"column:balance_date;type:date"`
	AutoDebit                    bool       `gorm:"column:auto_debit;not null"`
	AutoDebitSince               *time.Time `gorm:"column:auto_debit_since;type:date"`
	AutoDebitAccountID           *uuid.UUID `gorm:"column:auto_debit_account_id;type:uuid"`
	Notes                        *string    `gorm:"type:text"`
	CreatedAt                    time.Time  `gorm:"not null"`
	UpdatedAt                    time.Time  `gorm:"not null"`
	DeletedAt                    gorm.DeletedAt
}

func (FinancingContractModel) TableName() string { return "finance_financing_contracts" }

func financingToModel(c *dom.FinancingContract) FinancingContractModel {
	return FinancingContractModel{
		ID:                           c.ID,
		WorkspaceID:                  c.WorkspaceID,
		AssetType:                    string(c.AssetType),
		AssetID:                      c.AssetID,
		GroupID:                      c.GroupID,
		SupplierID:                   c.SupplierID,
		Description:                  c.Description,
		ContractNumber:               c.ContractNumber,
		PurchaseDate:                 c.PurchaseDate,
		FinancedCents:                c.FinancedCents,
		DownPaymentCents:             c.DownPaymentCents,
		TermMonths:                   c.TermMonths,
		NominalRatePPM:               c.NominalRatePPM,
		EffectiveRatePPM:             c.EffectiveRatePPM,
		EffectiveRelationshipRatePPM: c.EffectiveRelationshipRatePPM,
		InstallmentCents:             c.InstallmentCents,
		ReferenceDate:                c.ReferenceDate,
		ReferenceInstallment:         c.ReferenceInstallment,
		FirstDueDate:                 c.FirstDueDate,
		BalanceCents:                 c.BalanceCents,
		BalanceDate:                  c.BalanceDate,
		AutoDebit:                    c.AutoDebit,
		AutoDebitSince:               c.AutoDebitSince,
		AutoDebitAccountID:           c.AutoDebitAccountID,
		Notes:                        c.Notes,
		CreatedAt:                    c.CreatedAt,
		UpdatedAt:                    c.UpdatedAt,
	}
}

func modelToFinancing(m *FinancingContractModel) dom.FinancingContract {
	return dom.FinancingContract{
		ID:                           m.ID,
		WorkspaceID:                  m.WorkspaceID,
		AssetType:                    dom.AssetType(m.AssetType),
		AssetID:                      m.AssetID,
		GroupID:                      m.GroupID,
		SupplierID:                   m.SupplierID,
		Description:                  m.Description,
		ContractNumber:               m.ContractNumber,
		PurchaseDate:                 m.PurchaseDate,
		FinancedCents:                m.FinancedCents,
		DownPaymentCents:             m.DownPaymentCents,
		TermMonths:                   m.TermMonths,
		NominalRatePPM:               m.NominalRatePPM,
		EffectiveRatePPM:             m.EffectiveRatePPM,
		EffectiveRelationshipRatePPM: m.EffectiveRelationshipRatePPM,
		InstallmentCents:             m.InstallmentCents,
		ReferenceDate:                m.ReferenceDate,
		ReferenceInstallment:         m.ReferenceInstallment,
		FirstDueDate:                 m.FirstDueDate,
		BalanceCents:                 m.BalanceCents,
		BalanceDate:                  m.BalanceDate,
		AutoDebit:                    m.AutoDebit,
		AutoDebitSince:               m.AutoDebitSince,
		AutoDebitAccountID:           m.AutoDebitAccountID,
		Notes:                        m.Notes,
		CreatedAt:                    m.CreatedAt,
		UpdatedAt:                    m.UpdatedAt,
	}
}

type FinancingContractRepository struct {
	db *gorm.DB
}

func NewFinancingContractRepository(db *gorm.DB) *FinancingContractRepository {
	return &FinancingContractRepository{db: db}
}

func (r *FinancingContractRepository) Create(ctx context.Context, c *dom.FinancingContract) error {
	m := financingToModel(c)
	return mapFinanceErr(r.db.WithContext(ctx).Create(&m).Error)
}

// Update grava o contrato inteiro (Save): os campos opcionais precisam poder
// voltar a NULL, o que Updates com struct não faz.
func (r *FinancingContractRepository) Update(ctx context.Context, c *dom.FinancingContract) error {
	m := financingToModel(c)
	res := r.db.WithContext(ctx).Model(&FinancingContractModel{}).
		Where("id = ? AND workspace_id = ?", c.ID, c.WorkspaceID).
		Select("*").Omit("id", "workspace_id", "created_at", "deleted_at").
		Updates(&m)
	if res.Error != nil {
		return mapFinanceErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return dom.ErrNotFound
	}
	return nil
}

func (r *FinancingContractRepository) GetByID(ctx context.Context, workspaceID, id uuid.UUID) (*dom.FinancingContract, error) {
	var m FinancingContractModel
	if err := r.db.WithContext(ctx).Where("id = ? AND workspace_id = ?", id, workspaceID).First(&m).Error; err != nil {
		return nil, mapFinanceErr(err)
	}
	c := modelToFinancing(&m)
	return &c, nil
}

func (r *FinancingContractRepository) ListByAsset(ctx context.Context, workspaceID uuid.UUID, assetType dom.AssetType, assetID uuid.UUID) ([]dom.FinancingContract, error) {
	var ms []FinancingContractModel
	err := r.db.WithContext(ctx).
		Where("workspace_id = ? AND asset_type = ? AND asset_id = ?", workspaceID, string(assetType), assetID).
		Order("created_at DESC").Find(&ms).Error
	if err != nil {
		return nil, mapFinanceErr(err)
	}
	return modelsToFinancings(ms), nil
}

func (r *FinancingContractRepository) ListAutoDebit(ctx context.Context) ([]dom.FinancingContract, error) {
	var ms []FinancingContractModel
	if err := r.db.WithContext(ctx).Where("auto_debit = ?", true).Find(&ms).Error; err != nil {
		return nil, mapFinanceErr(err)
	}
	return modelsToFinancings(ms), nil
}

func (r *FinancingContractRepository) DeleteByGroup(ctx context.Context, workspaceID, groupID uuid.UUID) error {
	return mapFinanceErr(r.db.WithContext(ctx).
		Where("workspace_id = ? AND group_id = ?", workspaceID, groupID).
		Delete(&FinancingContractModel{}).Error)
}

func modelsToFinancings(ms []FinancingContractModel) []dom.FinancingContract {
	out := make([]dom.FinancingContract, 0, len(ms))
	for i := range ms {
		out = append(out, modelToFinancing(&ms[i]))
	}
	return out
}
