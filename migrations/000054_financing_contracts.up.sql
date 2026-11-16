-- Contrato de financiamento de um bem (imóvel ou veículo).
--
-- As parcelas continuam sendo lançamentos comuns (financial_entries), geradas
-- pelo parcelamento que já existe; quitação, renegociação e linhagem seguem
-- operando sobre a série. O contrato guarda o que a série não sabe: os dados
-- do banco (número, valor financiado, prazo, taxas), a posição de referência
-- usada para montar as parcelas e o saldo devedor informado — que NÃO é a
-- soma das parcelas em aberto (essa soma carrega juros futuros e correção).
--
-- group_id é o recurrence_group_id da série RAIZ. Depois de uma renegociação
-- a série vigente é outra, alcançada pela linhagem a partir da raiz.
-- Sem FK física para o bem, como em financial_entries (migração 000050).
--
-- Taxas anuais em ppm da fração (4,5939% a.a. = 45939): basis points não
-- comportam as quatro casas que os bancos publicam.
CREATE TABLE IF NOT EXISTS finance_financing_contracts (
    id                              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id                    UUID         NOT NULL,
    asset_type                      VARCHAR(20)  NOT NULL,
    asset_id                        UUID         NOT NULL,
    group_id                        UUID         NOT NULL,
    supplier_id                     UUID         REFERENCES suppliers (id) ON DELETE SET NULL,
    description                     VARCHAR(255) NOT NULL,
    contract_number                 VARCHAR(60),
    purchase_date                   DATE,
    financed_cents                  BIGINT,
    down_payment_cents              BIGINT,
    term_months                     INT          NOT NULL,
    nominal_rate_ppm                INT,
    effective_rate_ppm              INT,
    effective_relationship_rate_ppm INT,
    installment_cents               BIGINT       NOT NULL,
    reference_date                  DATE         NOT NULL,
    reference_installment           INT          NOT NULL,
    first_due_date                  DATE         NOT NULL,
    balance_cents                   BIGINT,
    balance_date                    DATE,
    auto_debit                      BOOLEAN      NOT NULL DEFAULT false,
    auto_debit_since                DATE,
    auto_debit_account_id           UUID         REFERENCES finance_accounts (id) ON DELETE SET NULL,
    notes                           TEXT,
    created_at                      TIMESTAMP    NOT NULL DEFAULT now(),
    updated_at                      TIMESTAMP    NOT NULL DEFAULT now(),
    deleted_at                      TIMESTAMP,
    CONSTRAINT chk_financing_contracts_asset_type CHECK (asset_type IN ('vehicle', 'property')),
    CONSTRAINT chk_financing_contracts_term CHECK (term_months BETWEEN 2 AND 600),
    CONSTRAINT chk_financing_contracts_reference
        CHECK (reference_installment BETWEEN 1 AND term_months)
);

CREATE INDEX IF NOT EXISTS idx_financing_contracts_asset
    ON finance_financing_contracts (workspace_id, asset_type, asset_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_financing_contracts_group
    ON finance_financing_contracts (workspace_id, group_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_financing_contracts_auto_debit
    ON finance_financing_contracts (auto_debit) WHERE auto_debit AND deleted_at IS NULL;
