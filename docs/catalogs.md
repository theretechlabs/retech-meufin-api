# Catálogo unificado de enums — `GET /api/v1/catalogs`

## O que é

Uma chamada, autenticada e cacheável, que devolve **todos os enums de domínio**
da API com rótulo humano. É a fonte única dos selects e chips do admin: o admin
não mantém mais constantes `*_OPTIONS` / `*_LABEL`; ele pede o catálogo uma vez
por sessão e resolve `slug → name` a partir daí.

Regra (CLAUDE.md): enum vive em código Go (`type X string` + constantes), nunca
em tabela; o rótulo também é do domínio. **Novo enum = novo item no catálogo**,
não uma constante nova no admin.

Os endpoints antigos (`/finance/discount-reasons`, `/finance/cancel-reasons`,
`/finance/card-brands`, `/home-safety/catalog`) continuam existindo; os três do
financeiro são servidos pelo mesmo dado (`dom.DiscountReasons`,
`dom.CancelReasons`, `dom.CardBrands`), sem duplicação.

## Contrato

```
GET /api/v1/catalogs
Cookie: meufin_session=…          (só autenticação; sem RequireModule)
If-None-Match: "<version>"        (opcional)

200 OK
ETag: "<version>"
Cache-Control: private, max-age=3600
{
  "version": "3f9a1c0b7d2e4a65",          // sha256 (16 hex) do conteúdo
  "catalogs": {
    "finance.payment_method": [
      { "slug": "pix", "name": "Pix" },
      { "slug": "compensacao", "name": "Compensação (sem caixa)",
        "description": "Gravada pelo sistema na quitação ou venda de uma troca de bem…" }
    ],
    "health.specialty": [ … ],
    …
  }
}

304 Not Modified   quando If-None-Match bate com o ETag (aceita W/, lista e *)
```

- `slug` é o valor que o admin envia e o banco grava; `name` e `description`
  são só apresentação. `description` é omitida quando vazia.
- A ordem dos itens é a ordem de exibição.
- `version` muda sempre que qualquer catálogo muda (é o hash do JSON canônico).
  O admin usa `version` como chave de cache; o browser usa o ETag.

## Chaves servidas (55)

| Módulo | Chaves |
|---|---|
| finance | `income_source_kind`, `income_type`, `recurrence`, `entry_status`, `account_kind`, `payment_method`, `attachment_type`, `supplier_category`, `supplier_billing_type`, `renegotiation_kind`, `card_bank`, `bank_account_type`, `asset_type`, `discount_reason`, `cancel_reason`, `card_brand` |
| education | `stage`, `shift`, `list_status`, `item_category` |
| health | `relationship`, `gender`, `lab_kind`, `plan_type`, `appointment_kind`, `specialty`, `appointment_status`, `member_doc_type`, `plan_doc_type`, `document_type`, `exam_source_type`, `exam_result_status`, `marker_category` |
| homesafety | `category`, `risk_type`, `priority`, `event_type`, `status` |
| patrimony | `property_type`, `property_doc_type`, `asset_type`, `tax_type`, `tax_status` |
| vehicle | `fuel_type`, `vehicle_status`, `fipe_vehicle_type`, `maintenance_type`, `maintenance_status`, `os_item_type`, `os_item_category`, `payment_method`, `alert_status`, `schedule_alert_status` |
| warranty | `category`, `status`, `doc_type` |

A chave completa é `<módulo>.<enum>` (ex.: `vehicle.fuel_type`). A lista é
espelhada no teste `internal/application/catalog/service_test.go` (`adminKeys`)
e no `CatalogKey` do admin (`src/lib/catalogs/api.ts`).

### Catálogos de campo livre

Alguns campos são gravados como texto livre no banco (não têm `type X string`
em Go). Para esses, o catálogo é a **lista curada que o admin oferece**, e o
comentário no `catalog.go` do módulo diz isso:

`finance.income_type`, `finance.card_bank`, `finance.bank_account_type`,
`health.gender`, `health.marker_category`, `vehicle.fipe_vehicle_type`,
`vehicle.maintenance_type` (slug = o próprio rótulo, dados legados),
`vehicle.payment_method` (lista própria da frota, diferente da do financeiro).

## Onde fica cada coisa

| Camada | Arquivo | Papel |
|---|---|---|
| domínio (folha) | `internal/domain/catalog/catalog.go` | `Item {Slug, Name, Description}`, `Catalog {Key, Items}`, helper `FromLabels` |
| domínio (módulo) | `internal/domain/<módulo>/catalog.go` | `func Catalogs() []catalog.Catalog` com os enums do módulo: slice ordenado + mapa de rótulos |
| aplicação | `internal/application/catalog/service.go` | `registry()` concatena os módulos; `New()` valida (chave duplicada, item vazio, slug/nome vazio, slug duplicado) e calcula `version`; resposta cacheada em memória |
| HTTP | `internal/interfaces/http/handlers/catalog_handler.go` | `GET /api/v1/catalogs`, ETag / 304 |
| composição | `cmd/api/main.go` | `CatalogService: appcat.MustNew()` — catálogo quebrado impede a API de subir |

## Como adicionar um valor a um enum existente

1. Adicione a constante no arquivo do enum (ex.: `PaymentMethodX PaymentMethod = "x"`
   em `internal/domain/finance/entry.go`) e no `ValidX`, se houver.
2. No `internal/domain/<módulo>/catalog.go`, acrescente a constante no slice
   `<enum>Order` (na posição em que deve aparecer) **e** o rótulo no mapa
   `<enum>Labels`.
3. `go test ./internal/application/catalog/...`: falta de rótulo ou slug fora do
   validador derruba o teste (e `MustNew` derrubaria a API na partida).
4. Nada a fazer no admin: ele lê o catálogo.

## Como adicionar um enum novo

1. `type X string` + constantes no módulo (regra do CLAUDE.md).
2. Em `internal/domain/<módulo>/catalog.go`: slice ordenado, mapa de rótulos e
   uma nova entrada `{Key: "<módulo>.<enum>", Items: catalog.FromLabels(...)}`
   em `Catalogs()`.
3. Adicione a chave em `adminKeys` no `service_test.go` **e** no `CatalogKey`
   do admin. O teste `TestNew_ServesEveryAdminKey` acusa chave servida sem par
   no admin e chave do admin sem catálogo.

## Como adicionar um módulo novo

Crie `internal/domain/<módulo>/catalog.go` com `Catalogs()` e acrescente uma
linha em `registry()` (`internal/application/catalog/service.go`).

## Divergências conhecidas (admin em `main` × enum Go)

Registradas aqui para a migração do admin tratar:

- `health.exam_source_type`: o admin oferecia `manual | document | import`; o
  enum Go é `manual | pdf | image | ocr | llm` (a importação grava `llm`). O
  catálogo serve o enum Go.
- `health.exam_result_status`: o admin oferecia `draft | reviewed | final`; o
  enum Go é `draft | processing | extracted | reviewed | failed` (a importação
  grava `extracted`). O catálogo serve o enum Go.
- `finance.supplier_category`: o Go tem quatro categorias automotivas
  (`auto_center`, `mecanica`, `posto_combustivel`, `loja_pecas`) que o admin
  não listava; entram no catálogo antes de `outros`.
- `finance.payment_method`: `compensacao` entra no catálogo (o admin precisa do
  rótulo para exibir lançamentos existentes), com `description` avisando que só
  o sistema a grava. O admin deve excluí-la do select de liquidação manual.
