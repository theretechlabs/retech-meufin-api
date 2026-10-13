package finance

import (
	"testing"
	"time"
)

func d(y int, m time.Month, day int) time.Time { return time.Date(y, m, day, 0, 0, 0, 0, time.UTC) }

func TestDaysLate(t *testing.T) {
	due := d(2026, 10, 10)
	cases := []struct {
		paid time.Time
		want int
	}{
		{d(2026, 10, 9), 0},
		{d(2026, 10, 10), 0},
		{time.Date(2026, 10, 10, 23, 59, 0, 0, time.UTC), 0},
		{d(2026, 10, 11), 1},
		{d(2026, 11, 9), 30},
		// Hora não conta: 12h do dia seguinte ainda é 1 dia.
		{time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC), 1},
	}
	for _, c := range cases {
		if got := DaysLate(due, c.paid); got != c.want {
			t.Errorf("DaysLate(%s, %s) = %d, quer %d", due.Format("2006-01-02"), c.paid.Format("2006-01-02T15:04"), got, c.want)
		}
	}
}

func TestQuoteLateChargesDefaultTerms(t *testing.T) {
	e := FinancialEntry{AmountCents: 100_000, DueDate: d(2026, 10, 10)}

	// Sem atraso: tudo zero, mas os termos aparecem (a UI mostra o que valeria).
	q := e.QuoteLateCharges(d(2026, 10, 10))
	if q.DaysLate != 0 || q.TotalCents != 0 || q.LateFeeCents != 0 || q.LateInterestCents != 0 {
		t.Fatalf("sem atraso deve zerar, veio %+v", q)
	}
	if q.Source != LateTermsDefault || q.LateFeeBP != DefaultLateFeeBP || q.LateInterestMonthlyBP != DefaultLateInterestMonthlyBP {
		t.Fatalf("termos padrão esperados, veio %+v", q)
	}

	// 15 dias: multa 2% = 2.000; juros 1% a.m. × 15/30 = 500.
	q = e.QuoteLateCharges(d(2026, 10, 25))
	if q.DaysLate != 15 || q.LateFeeCents != 2_000 || q.LateInterestCents != 500 || q.TotalCents != 2_500 {
		t.Fatalf("15 dias: quer multa 2000 / juros 500, veio %+v", q)
	}

	// 30 dias: juros cheios de 1% = 1.000.
	q = e.QuoteLateCharges(d(2026, 11, 9))
	if q.LateInterestCents != 1_000 || q.TotalCents != 3_000 {
		t.Fatalf("30 dias: quer juros 1000, veio %+v", q)
	}
}

func TestQuoteLateChargesEntryTerms(t *testing.T) {
	fee, interest := 1_000, 300 // 10% multa, 3% a.m.
	e := FinancialEntry{AmountCents: 50_000, DueDate: d(2026, 1, 31), LateFeeBP: &fee, LateInterestMonthlyBP: &interest}
	q := e.QuoteLateCharges(d(2026, 2, 10)) // 10 dias
	if q.Source != LateTermsEntry {
		t.Fatalf("source = %s, quer entry", q.Source)
	}
	// multa 10% = 5.000; juros 3% × 10/30 = 1% = 500.
	if q.LateFeeCents != 5_000 || q.LateInterestCents != 500 || q.TotalCents != 5_500 {
		t.Fatalf("quer 5000/500, veio %+v", q)
	}
}

func TestQuoteLateChargesOnlyOneTermMeansOtherIsZero(t *testing.T) {
	fee := 200
	e := FinancialEntry{AmountCents: 10_000, DueDate: d(2026, 3, 1), LateFeeBP: &fee}
	q := e.QuoteLateCharges(d(2026, 3, 31))
	if q.LateInterestMonthlyBP != 0 || q.LateInterestCents != 0 {
		t.Fatalf("juros não informados devem valer zero (não o padrão), veio %+v", q)
	}
	if q.LateFeeCents != 200 {
		t.Fatalf("multa 2%% de 10000 = 200, veio %d", q.LateFeeCents)
	}
}

func TestQuoteLateChargesRounding(t *testing.T) {
	// 1 dia de 1% a.m. sobre R$ 1.000,00 = 100000 × 100 × 1 / 300000 = 33,33 → 33.
	e := FinancialEntry{AmountCents: 100_000, DueDate: d(2026, 5, 5)}
	q := e.QuoteLateCharges(d(2026, 5, 6))
	if q.LateInterestCents != 33 {
		t.Fatalf("quer 33 centavos, veio %d", q.LateInterestCents)
	}
	// 2 dias = 66,67 → 67 (meio para cima).
	q = e.QuoteLateCharges(d(2026, 5, 7))
	if q.LateInterestCents != 67 {
		t.Fatalf("quer 67 centavos, veio %d", q.LateInterestCents)
	}
}

func TestValidateLateTermsAndCharges(t *testing.T) {
	base := func() FinancialEntry {
		return FinancialEntry{WorkspaceID: [16]byte{1}, Kind: KindDebit, Status: StatusPrevista, AmountCents: 100, DueDate: d(2026, 1, 1), Type: strp("moradia")}
	}
	e := base()
	bad := 20_000
	e.LateFeeBP = &bad
	if err := e.Validate(); err == nil {
		t.Fatal("multa acima de 100% deve falhar")
	}
	e = base()
	neg := -1
	e.LateInterestMonthlyBP = &neg
	if err := e.Validate(); err == nil {
		t.Fatal("juros negativos devem falhar")
	}
	e = base()
	fee := int64(10)
	e.LateFeeCents = &fee
	if err := e.Validate(); err == nil {
		t.Fatal("encargo pago em lançamento previsto deve falhar")
	}
	e.Status = StatusRealizada
	if err := e.Validate(); err != nil {
		t.Fatalf("encargo pago em realizado é válido: %v", err)
	}
}

func strp(s string) *string { return &s }
