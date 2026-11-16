package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	app "github.com/retechfin/retechfin-api/internal/application/finance"
	dom "github.com/retechfin/retechfin-api/internal/domain/finance"
	"github.com/retechfin/retechfin-api/internal/interfaces/http/errrespond"
	"github.com/retechfin/retechfin-api/internal/interfaces/http/middleware"
)

// FinancingHandler expõe os contratos de financiamento de bens.
type FinancingHandler struct {
	svc *app.FinancingService
}

func NewFinancingHandler(svc *app.FinancingService) *FinancingHandler {
	return &FinancingHandler{svc: svc}
}

func mapFinancing(c *dom.FinancingContract) gin.H {
	return gin.H{
		"id":                              c.ID,
		"asset_type":                      string(c.AssetType),
		"asset_id":                        c.AssetID,
		"group_id":                        c.GroupID,
		"supplier_id":                     c.SupplierID,
		"description":                     c.Description,
		"contract_number":                 c.ContractNumber,
		"purchase_date":                   fmtDatePtr(c.PurchaseDate),
		"financed_cents":                  c.FinancedCents,
		"down_payment_cents":              c.DownPaymentCents,
		"term_months":                     c.TermMonths,
		"nominal_rate_ppm":                c.NominalRatePPM,
		"effective_rate_ppm":              c.EffectiveRatePPM,
		"effective_relationship_rate_ppm": c.EffectiveRelationshipRatePPM,
		"installment_cents":               c.InstallmentCents,
		"reference_date":                  c.ReferenceDate.Format(entryDateLayout),
		"reference_installment":           c.ReferenceInstallment,
		"remaining_months":                c.RemainingMonths(),
		"first_due_date":                  c.FirstDueDate.Format(entryDateLayout),
		"balance_cents":                   c.BalanceCents,
		"balance_date":                    fmtDatePtr(c.BalanceDate),
		"auto_debit":                      c.AutoDebit,
		"auto_debit_since":                fmtDatePtr(c.AutoDebitSince),
		"auto_debit_account_id":           c.AutoDebitAccountID,
		"notes":                           c.Notes,
		"created_at":                      c.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at":                      c.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

// financingTermsJSON são os dados do contrato editáveis depois do cadastro.
type financingTermsJSON struct {
	ContractNumber               *string    `json:"contract_number"`
	PurchaseDate                 *string    `json:"purchase_date"`
	FinancedCents                *int64     `json:"financed_cents"`
	DownPaymentCents             *int64     `json:"down_payment_cents"`
	NominalRatePPM               *int       `json:"nominal_rate_ppm"`
	EffectiveRatePPM             *int       `json:"effective_rate_ppm"`
	EffectiveRelationshipRatePPM *int       `json:"effective_relationship_rate_ppm"`
	BalanceCents                 *int64     `json:"balance_cents"`
	BalanceDate                  *string    `json:"balance_date"`
	AutoDebit                    bool       `json:"auto_debit"`
	AutoDebitAccountID           *uuid.UUID `json:"auto_debit_account_id"`
	Notes                        *string    `json:"notes"`
}

func (j financingTermsJSON) toTerms() (app.FinancingTerms, string) {
	purchase, msg := parseDatePtr(j.PurchaseDate, "purchase_date")
	if msg != "" {
		return app.FinancingTerms{}, msg
	}
	balanceDate, msg := parseDatePtr(j.BalanceDate, "balance_date")
	if msg != "" {
		return app.FinancingTerms{}, msg
	}
	return app.FinancingTerms{
		ContractNumber:               trimPtr(j.ContractNumber),
		PurchaseDate:                 purchase,
		FinancedCents:                j.FinancedCents,
		DownPaymentCents:             j.DownPaymentCents,
		NominalRatePPM:               j.NominalRatePPM,
		EffectiveRatePPM:             j.EffectiveRatePPM,
		EffectiveRelationshipRatePPM: j.EffectiveRelationshipRatePPM,
		BalanceCents:                 j.BalanceCents,
		BalanceDate:                  balanceDate,
		AutoDebit:                    j.AutoDebit,
		AutoDebitAccountID:           j.AutoDebitAccountID,
		Notes:                        trimPtr(j.Notes),
	}, ""
}

type financingCreateJSON struct {
	AssetType            string     `json:"asset_type" binding:"required"`
	AssetID              uuid.UUID  `json:"asset_id" binding:"required"`
	SupplierID           *uuid.UUID `json:"supplier_id"`
	Description          string     `json:"description" binding:"required"`
	TermMonths           int        `json:"term_months" binding:"required"`
	InstallmentCents     int64      `json:"installment_cents" binding:"required"`
	ReferenceDate        string     `json:"reference_date" binding:"required"`
	ReferenceInstallment int        `json:"reference_installment" binding:"required"`
	SettlePast           bool       `json:"settle_past"`
	financingTermsJSON
}

// Create responde POST /finance/financings — registra o contrato e gera a
// série inteira de parcelas vinculada ao bem.
func (h *FinancingHandler) Create(c *gin.Context) {
	ws, ok := middleware.WorkspaceID(c)
	if !ok {
		errrespond.Message(c, http.StatusBadRequest, errrespond.CodeWorkspaceRequired, "workspace inválido")
		return
	}
	var body financingCreateJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		errrespond.Message(c, http.StatusBadRequest, errrespond.CodeBadRequest, "JSON inválido")
		return
	}
	ref, err := time.Parse(entryDateLayout, body.ReferenceDate)
	if err != nil {
		errrespond.Message(c, http.StatusBadRequest, errrespond.CodeBadRequest, "reference_date inválida (use YYYY-MM-DD)")
		return
	}
	terms, msg := body.toTerms()
	if msg != "" {
		errrespond.Message(c, http.StatusBadRequest, errrespond.CodeBadRequest, msg)
		return
	}
	res, err := h.svc.Create(c.Request.Context(), app.CreateFinancingInput{
		WorkspaceID:          ws,
		AssetType:            dom.AssetType(body.AssetType),
		AssetID:              body.AssetID,
		SupplierID:           body.SupplierID,
		Description:          body.Description,
		TermMonths:           body.TermMonths,
		InstallmentCents:     body.InstallmentCents,
		ReferenceDate:        ref,
		ReferenceInstallment: body.ReferenceInstallment,
		SettlePast:           body.SettlePast,
		FinancingTerms:       terms,
	})
	if err != nil {
		errrespond.Write(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"contract": mapFinancing(res.Contract),
		"created":  res.Created,
		"settled":  res.Settled,
	})
}

func (h *FinancingHandler) contractID(c *gin.Context) (ws, id uuid.UUID, ok bool) {
	ws, ok = middleware.WorkspaceID(c)
	if !ok {
		errrespond.Message(c, http.StatusBadRequest, errrespond.CodeWorkspaceRequired, "workspace inválido")
		return ws, id, false
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		errrespond.Message(c, http.StatusBadRequest, errrespond.CodeBadRequest, "id inválido")
		return ws, id, false
	}
	return ws, id, true
}

// Update responde PUT /finance/financings/:id — corrige os dados do contrato
// (número, valores, taxas, saldo devedor informado, débito automático).
func (h *FinancingHandler) Update(c *gin.Context) {
	ws, id, ok := h.contractID(c)
	if !ok {
		return
	}
	var body financingTermsJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		errrespond.Message(c, http.StatusBadRequest, errrespond.CodeBadRequest, "JSON inválido")
		return
	}
	terms, msg := body.toTerms()
	if msg != "" {
		errrespond.Message(c, http.StatusBadRequest, errrespond.CodeBadRequest, msg)
		return
	}
	contract, err := h.svc.Update(c.Request.Context(), ws, id, terms)
	if err != nil {
		errrespond.Write(c, err)
		return
	}
	c.JSON(http.StatusOK, mapFinancing(contract))
}

type financingReadjustJSON struct {
	InstallmentCents int64  `json:"installment_cents" binding:"required"`
	FromDate         string `json:"from_date" binding:"required"`
}

// Readjust responde POST /finance/financings/:id/readjust — novo valor da
// prestação para as parcelas em aberto a partir de from_date.
func (h *FinancingHandler) Readjust(c *gin.Context) {
	ws, id, ok := h.contractID(c)
	if !ok {
		return
	}
	var body financingReadjustJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		errrespond.Message(c, http.StatusBadRequest, errrespond.CodeBadRequest, "JSON inválido")
		return
	}
	from, err := time.Parse(entryDateLayout, body.FromDate)
	if err != nil {
		errrespond.Message(c, http.StatusBadRequest, errrespond.CodeBadRequest, "from_date inválida (use YYYY-MM-DD)")
		return
	}
	contract, updated, err := h.svc.Readjust(c.Request.Context(), ws, id, body.InstallmentCents, from)
	if err != nil {
		errrespond.Write(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"contract": mapFinancing(contract), "updated": updated})
}

// Delete responde DELETE /finance/financings/:id — remove o contrato e a
// série inteira de parcelas.
func (h *FinancingHandler) Delete(c *gin.Context) {
	ws, id, ok := h.contractID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), ws, id); err != nil {
		errrespond.Write(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
