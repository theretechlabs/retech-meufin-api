DROP INDEX IF EXISTS idx_financial_entries_late_charges;
ALTER TABLE financial_entries DROP COLUMN IF EXISTS late_interest_cents;
ALTER TABLE financial_entries DROP COLUMN IF EXISTS late_fee_cents;
ALTER TABLE financial_entries DROP COLUMN IF EXISTS late_interest_monthly_bp;
ALTER TABLE financial_entries DROP COLUMN IF EXISTS late_fee_bp;
