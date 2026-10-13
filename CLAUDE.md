# retech-meufin-api — regras de desenvolvimento

## Regra principal: reutilizar, não duplicar

1. **Uma implementação por operação.** Antes de escrever um fluxo novo (liquidar, cancelar, importar…), procure o que já existe no `application/` e estenda. Dois caminhos para a mesma operação (ex.: `confirm` e `settle` com regras divergentes) é bug estrutural: converge num núcleo e faça os dois delegarem.
2. **Ao reaproveitar, apague o que foi substituído.** Código sem uso não fica no repositório. Handlers duplicam cada vez menos: HTTP comum vira helper (`postMessages`, `pagination`, `boolQuery`).
3. **Validar premissas no código antes de executar.** Não assumir que o admin, o auth ou o banco se comportam como a memória diz: ler o código e, se houver dúvida, testar.

## Listas e enums (contrato com o admin)

- **Enum de domínio** (status, kind, recorrência, forma de pagamento, tipo de documento…): vive em código Go no `internal/domain/<módulo>` (`type X string` + `ValidX`), **nunca em tabela**. Rótulo humano também é do domínio: a API publica os enums com `slug`/`name`/`description` num endpoint de catálogo **não paginado**, cacheável, uma chamada para o admin (padrão: `/finance/discount-reasons`, `/finance/cancel-reasons`, `/finance/card-brands`, `/home-safety/catalog`). Novo enum = novo item no catálogo, não uma constante nova no admin.
- **Lista de entidade do workspace** (cartões, contas, fornecedores, categorias, membros…): endpoint paginado com `limit`/`offset` (`pagination(c)`), `query` para busca e `active` (`boolQuery`). O admin pede `active=true` para montar seletores e nunca baixa a lista inteira para filtrar no browser quando ela pode crescer.
- Inativo/arquivado continua existindo para histórico e para resolver nome em registros antigos; só não entra em lista de opção.

## Outras convenções

- Dinheiro em centavos (`int64`); percentuais em basis points; datas de negócio `YYYY-MM-DD` (`entryDateLayout`); `paid_at` é o eixo do caixa — toda liquidação grava `paid_at` e `paid_amount_cents`.
- Erros de regra: `&dom.ValidationError{Msg: "..."}` com mensagem em português que o admin mostra como está.
- Integrações externas (Anthropic, Infosimples, auth) ficam em `internal/infrastructure/*` atrás de uma porta; o `application/` não conhece HTTP de terceiros.
- Chamadas à Anthropic: `tool_choice` sempre `auto` (modelos 5.x rejeitam `tool`/`any`); `max_tokens` inclui o thinking; `output_config.effort` por caminho.
- Antes de commitar: `gofmt -l internal/`, `go vet ./...`, `go test ./...`.
- Git: branch nova a partir da `main` atualizada; nunca commitar em branch de PR já mergeada.
