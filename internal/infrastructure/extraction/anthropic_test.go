package extraction

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeAnthropic simula a Messages API: grava cada requisição e devolve as
// respostas na ordem informada.
type fakeAnthropic struct {
	t         *testing.T
	requests  []map[string]any
	responses []string
}

func (f *fakeAnthropic) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			f.t.Fatalf("payload inválido: %v", err)
		}
		f.requests = append(f.requests, req)
		i := len(f.requests) - 1
		if i >= len(f.responses) {
			f.t.Fatalf("requisição inesperada #%d", i+1)
		}
		w.Header().Set("content-type", "application/json")
		_, _ = w.Write([]byte(f.responses[i]))
	}
}

func newTestExtractor(t *testing.T, fake *fakeAnthropic) *AnthropicExtractor {
	t.Helper()
	srv := httptest.NewServer(fake.handler())
	t.Cleanup(srv.Close)
	return NewAnthropicExtractor(Config{Provider: ProviderAnthropic, APIKey: "k", BaseURL: srv.URL, Model: "claude-test"})
}

const textOnlyResponse = `{"type":"message","model":"claude-test","stop_reason":"end_turn","content":[{"type":"text","text":"Segue a chave: 123"}]}`

func toolUseResponse(tool string, input string) string {
	return `{"type":"message","model":"claude-test","stop_reason":"tool_use","content":[{"type":"tool_use","name":"` + tool + `","input":` + input + `}]}`
}

// Claude 5.x rejeita tool_choice "tool"/"any": toda chamada precisa sair com
// "auto" (e sem name), em todos os caminhos.
func TestToolChoiceIsAutoEverywhere(t *testing.T) {
	fake := &fakeAnthropic{t: t, responses: []string{
		toolUseResponse(keyReaderToolName, `{"chave_acesso":"1"}`),
		toolUseResponse(categorizeToolName, `{"items":[]}`),
		toolUseResponse(invoiceToolName, `{"items":[]}`),
	}}
	a := newTestExtractor(t, fake)
	ctx := context.Background()

	if _, err := a.ReadFiscalKey(ctx, []byte("img"), "image/png"); err != nil {
		t.Fatalf("ReadFiscalKey: %v", err)
	}
	if _, err := a.Categorize(ctx, CategorizeInput{Descriptions: []string{"x"}}); err != nil {
		t.Fatalf("Categorize: %v", err)
	}
	p := CreditCardInvoiceProfile()
	if _, err := a.Extract(ctx, ExtractInput{InputType: InputTypePDF, Content: []byte("pdf"), Profile: &p}); err != nil {
		t.Fatalf("Extract: %v", err)
	}

	if len(fake.requests) != 3 {
		t.Fatalf("quer 3 requisições, veio %d", len(fake.requests))
	}
	for i, req := range fake.requests {
		tc, _ := req["tool_choice"].(map[string]any)
		if tc["type"] != "auto" {
			t.Fatalf("req #%d: tool_choice.type = %v, quer auto", i+1, tc["type"])
		}
		if _, has := tc["name"]; has {
			t.Fatalf("req #%d: tool_choice não pode ter name com auto", i+1)
		}
		if tc["disable_parallel_tool_use"] != true {
			t.Fatalf("req #%d: quer disable_parallel_tool_use=true", i+1)
		}
	}
	// Effort por caminho: chave e categorização em low; extração no default.
	wantEffort := []string{EffortLow, EffortLow, DefaultEffort}
	for i, req := range fake.requests {
		oc, _ := req["output_config"].(map[string]any)
		if oc["effort"] != wantEffort[i] {
			t.Fatalf("req #%d: output_config.effort = %v, quer %s", i+1, oc["effort"], wantEffort[i])
		}
	}
}

// "auto" não garante a chamada: resposta só com texto dispara UM reenvio com
// reforço, e o resultado do reenvio é usado.
func TestCallToolRetriesOnceWhenModelAnswersWithText(t *testing.T) {
	fake := &fakeAnthropic{t: t, responses: []string{
		textOnlyResponse,
		toolUseResponse(keyReaderToolName, `{"chave_acesso":" 42260779 "}`),
	}}
	a := newTestExtractor(t, fake)

	key, err := a.ReadFiscalKey(context.Background(), []byte("img"), "image/png")
	if err != nil {
		t.Fatalf("ReadFiscalKey: %v", err)
	}
	if key != "42260779" {
		t.Fatalf("chave = %q, quer 42260779", key)
	}
	if len(fake.requests) != 2 {
		t.Fatalf("quer exatamente 2 requisições (1 retry), veio %d", len(fake.requests))
	}
	// O reenvio acrescenta o reforço ao fim da mensagem do usuário.
	msgs := fake.requests[1]["messages"].([]any)
	content := msgs[len(msgs)-1].(map[string]any)["content"].([]any)
	last := content[len(content)-1].(map[string]any)
	if !strings.Contains(last["text"].(string), keyReaderToolName) {
		t.Fatalf("reforço do retry deve citar a ferramenta, veio %v", last["text"])
	}
	// A primeira requisição não tinha o reforço.
	first := fake.requests[0]["messages"].([]any)[0].(map[string]any)["content"].([]any)
	if len(first) != len(content)-1 {
		t.Fatalf("retry deve ter exatamente 1 bloco a mais: %d vs %d", len(first), len(content))
	}
}

// Dois textos seguidos: só um retry; Extract falha claro, sem loop.
func TestExtractFailsClearlyAfterRetry(t *testing.T) {
	fake := &fakeAnthropic{t: t, responses: []string{textOnlyResponse, textOnlyResponse}}
	a := newTestExtractor(t, fake)
	p := CreditCardInvoiceProfile()

	_, err := a.Extract(context.Background(), ExtractInput{InputType: InputTypePDF, Content: []byte("pdf"), Profile: &p})
	if err == nil || !strings.Contains(err.Error(), "sem dados estruturados") {
		t.Fatalf("quer erro 'sem dados estruturados', veio %v", err)
	}
	if len(fake.requests) != 2 {
		t.Fatalf("quer 2 requisições, veio %d", len(fake.requests))
	}
}

// Erro HTTP da API vem com tipo e mensagem, sem retry.
func TestPostMessagesSurfacesAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"type":"error","error":{"type":"invalid_request_error","message":"boom"}}`))
	}))
	defer srv.Close()
	a := NewAnthropicExtractor(Config{Provider: ProviderAnthropic, APIKey: "k", BaseURL: srv.URL})

	_, err := a.ReadFiscalKey(context.Background(), []byte("img"), "image/png")
	if err == nil || !strings.Contains(err.Error(), "keyreader: anthropic 400 (invalid_request_error): boom") {
		t.Fatalf("erro inesperado: %v", err)
	}
}
