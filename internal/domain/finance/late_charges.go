package finance

import "time"

// Encargos por atraso — juros e multa.
//
// O lançamento pode carregar os TERMOS (multa % e juros % a.m.) informados ao
// lançar e herdados pela série/parcelas. Sem termos, vale o padrão abaixo: o
// teto do CDC para contratos de consumo (multa 2%, art. 52 §1º) e a mora de
// 1% ao mês praticada em boletos, pro rata dia na base 30.
//
// Os encargos PAGOS ficam no lançamento (LateFeeCents/LateInterestCents) e
// viram indicador: "quanto a família perdeu em juros e multa no mês".
const (
	// DefaultLateFeeBP é a multa padrão em basis points (200 = 2%).
	DefaultLateFeeBP = 200
	// DefaultLateInterestMonthlyBP são os juros padrão ao mês em basis
	// points (100 = 1% a.m.).
	DefaultLateInterestMonthlyBP = 100
	// MaxLateChargeBP é o teto aceito para cada termo (10000 = 100%): acima
	// disso é erro de digitação, não contrato.
	MaxLateChargeBP = 10_000
)

// LateTermsSource diz de onde vieram os termos usados numa cotação.
type LateTermsSource string

const (
	LateTermsEntry   LateTermsSource = "entry"   // informados no lançamento
	LateTermsDefault LateTermsSource = "default" // padrão do domínio
)

// LateChargeQuote é a cotação de encargos para uma data de pagamento.
type LateChargeQuote struct {
	DueDate  time.Time
	PaidAt   time.Time
	DaysLate int
	// BaseCents é o valor sobre o qual os percentuais incidem (valor do
	// lançamento, antes de desconto).
	BaseCents             int64
	LateFeeBP             int
	LateInterestMonthlyBP int
	Source                LateTermsSource
	LateFeeCents          int64
	LateInterestCents     int64
	TotalCents            int64
}

// DaysLate conta dias corridos entre o vencimento e o pagamento, ignorando a
// hora. Pagamento até o vencimento (inclusive) = 0.
func DaysLate(due, paidAt time.Time) int {
	d := dateOnlyUTC(due)
	p := dateOnlyUTC(paidAt)
	if !p.After(d) {
		return 0
	}
	return int(p.Sub(d).Hours() / 24)
}

func dateOnlyUTC(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// LateTerms resolve os termos efetivos do lançamento: os informados ou o
// padrão do domínio quando nenhum dos dois foi preenchido.
func (e *FinancialEntry) LateTerms() (feeBP, interestBP int, source LateTermsSource) {
	if e.LateFeeBP == nil && e.LateInterestMonthlyBP == nil {
		return DefaultLateFeeBP, DefaultLateInterestMonthlyBP, LateTermsDefault
	}
	if e.LateFeeBP != nil {
		feeBP = *e.LateFeeBP
	}
	if e.LateInterestMonthlyBP != nil {
		interestBP = *e.LateInterestMonthlyBP
	}
	return feeBP, interestBP, LateTermsEntry
}

// QuoteLateCharges calcula multa e juros de mora para pagar o lançamento em
// paidAt. Sem atraso, tudo zero. Arredondamento meio-para-cima no centavo,
// independente em cada parcela (multa e juros), como o boleto faz.
func (e *FinancialEntry) QuoteLateCharges(paidAt time.Time) LateChargeQuote {
	feeBP, interestBP, source := e.LateTerms()
	q := LateChargeQuote{
		DueDate:               dateOnlyUTC(e.DueDate),
		PaidAt:                dateOnlyUTC(paidAt),
		DaysLate:              DaysLate(e.DueDate, paidAt),
		BaseCents:             e.AmountCents,
		LateFeeBP:             feeBP,
		LateInterestMonthlyBP: interestBP,
		Source:                source,
	}
	if q.DaysLate <= 0 || q.BaseCents <= 0 {
		return q
	}
	// multa = base × fee/10000
	q.LateFeeCents = roundDiv(q.BaseCents*int64(feeBP), 10_000)
	// juros = base × interest/10000 × dias/30
	q.LateInterestCents = roundDiv(q.BaseCents*int64(interestBP)*int64(q.DaysLate), 10_000*30)
	q.TotalCents = q.LateFeeCents + q.LateInterestCents
	return q
}

// roundDiv divide inteiros com arredondamento meio-para-cima (num >= 0).
func roundDiv(num, den int64) int64 {
	if num <= 0 {
		return 0
	}
	return (num + den/2) / den
}
