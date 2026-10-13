package extraction

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// KeyReaderPromptVersion versiona o prompt de leitura da chave.
const KeyReaderPromptVersion = "fiscal-key-v1"

const keyReaderToolName = "extrair_chave_nfce"

// FiscalKeyReader lê a CHAVE DE ACESSO (44 dígitos) impressa num cupom/nota a
// partir da imagem — fallback quando o QR Code está ilegível (a chave impressa
// costuma sobreviver melhor ao borrão/compressão que o QR).
type FiscalKeyReader interface {
	Enabled() bool
	ReadFiscalKey(ctx context.Context, content []byte, mimeType string) (string, error)
}

// NewKeyReader devolve o leitor Anthropic quando há credencial, senão um
// leitor desabilitado (no-op).
func NewKeyReader(cfg Config) FiscalKeyReader {
	if cfg.Model == "" {
		cfg.Model = DefaultModel
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	if cfg.Provider == ProviderAnthropic && cfg.APIKey != "" {
		return NewAnthropicExtractor(cfg)
	}
	return disabledKeyReader{}
}

type disabledKeyReader struct{}

func (disabledKeyReader) Enabled() bool { return false }
func (disabledKeyReader) ReadFiscalKey(context.Context, []byte, string) (string, error) {
	return "", ErrExtractionDisabled
}

const keyReaderSystemPrompt = `Você extrai a CHAVE DE ACESSO de um cupom/nota fiscal brasileiro (NFC-e/NF-e) a partir da imagem.
A chave tem EXATAMENTE 44 dígitos, geralmente impressa perto do QR Code (às vezes em grupos de 4, ex.: "4226 0779 2572 ...").
Regras:
- Transcreva SOMENTE os 44 dígitos, sem espaços nem outros caracteres.
- Não confunda com CNPJ, número do documento, protocolo ou valor.
- Se não houver uma chave de 44 dígitos legível, retorne "".
Use SEMPRE a ferramenta ` + keyReaderToolName + `.`

func keyReaderInputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"chave_acesso": map[string]any{"type": "string", "description": "44 dígitos, sem espaços; \"\" se ausente/ilegível"},
		},
		"required": []string{"chave_acesso"},
	}
}

type keyReaderOutput struct {
	ChaveAcesso string `json:"chave_acesso"`
}

// ReadFiscalKey implementa FiscalKeyReader no AnthropicExtractor.
func (a *AnthropicExtractor) ReadFiscalKey(ctx context.Context, content []byte, mimeType string) (string, error) {
	if !a.Enabled() {
		return "", ErrExtractionDisabled
	}
	if len(content) == 0 {
		return "", fmt.Errorf("keyreader: conteúdo vazio")
	}
	inputType := InputTypeImage
	if strings.EqualFold(mimeType, "application/pdf") {
		inputType = InputTypePDF
	}
	docBlock, err := buildDocumentBlock(ExtractInput{InputType: inputType, MimeType: mimeType, Content: content})
	if err != nil {
		return "", err
	}

	reqBody := anthropicRequest{
		Model: a.cfg.Model,
		// Em modelos com thinking adaptativo (Claude 5.x) o raciocínio consome
		// max_tokens junto com a resposta; 512 truncava antes da ferramenta.
		MaxTokens: 4096,
		System:    keyReaderSystemPrompt,
		Tools: []anthropicTool{{
			Name:        keyReaderToolName,
			Description: "Extrai a chave de acesso (44 dígitos) impressa no cupom/nota fiscal.",
			InputSchema: keyReaderInputSchema(),
		}},
		ToolChoice: autoToolChoice(),
		Messages: []anthropicMessage{{
			Role: "user",
			Content: []anthropicContentBlock{
				docBlock,
				{Type: "text", Text: "Extraia a chave de acesso de 44 dígitos impressa neste cupom."},
			},
		}},
	}

	input, _, _, err := a.callTool(ctx, "keyreader", reqBody, keyReaderToolName)
	if err != nil {
		return "", err
	}
	if len(input) == 0 {
		return "", nil
	}
	var out keyReaderOutput
	if err := json.Unmarshal(input, &out); err != nil {
		return "", fmt.Errorf("keyreader: decodificar tool: %w", err)
	}
	return strings.TrimSpace(out.ChaveAcesso), nil
}
