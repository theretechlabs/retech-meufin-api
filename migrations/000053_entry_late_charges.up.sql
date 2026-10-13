-- Encargos por atraso (juros e multa).
--
-- Termos do lançamento (informados ao lançar, herdados pela série/parcelas):
--   late_fee_bp              multa única sobre o valor, em basis points (200 = 2%)
--   late_interest_monthly_bp juros de mora ao mês, em basis points (100 = 1% a.m.),
--                            pro rata dia (base 30)
-- NULL = usa o padrão do domínio (internal/domain/finance/late_charges.go).
--
-- Encargos efetivamente pagos na liquidação (indicador):
--   late_fee_cents / late_interest_cents — somam ao valor pago; reabrir limpa.
ALTER TABLE financial_entries ADD COLUMN IF NOT EXISTS late_fee_bp INTEGER;
ALTER TABLE financial_entries ADD COLUMN IF NOT EXISTS late_interest_monthly_bp INTEGER;
ALTER TABLE financial_entries ADD COLUMN IF NOT EXISTS late_fee_cents BIGINT;
ALTER TABLE financial_entries ADD COLUMN IF NOT EXISTS late_interest_cents BIGINT;

-- Indicador: "quanto paguei de juros/multa no mês" agrega só quem tem encargo.
CREATE INDEX IF NOT EXISTS idx_financial_entries_late_charges
    ON financial_entries (workspace_id, paid_at)
    WHERE late_fee_cents IS NOT NULL OR late_interest_cents IS NOT NULL;
