## Rollout & Rollback Plan for LLM + Cuotas Features

This repo now has **feature flags** and a simple rollback protocol for:

- **LLM features** (natural language `/add` and `/review` AI analysis)
- **Installment expenses** (`/add_cuotas` and installment metadata)

Use this guide when preparing a deploy.

---

### 1. Feature Flags (env vars)

Set these in your environment (or `.env`) on the host where the bot runs:

- **`ENABLE_LLM_FEATURES`** (default: `true`)
  - Controls:
    - Natural-language expense parsing in regular messages
    - Registration of `/review` command (AI-powered review)
    - Whether the LLM client is even initialized
  - If `false`, the bot behaves like a pure non-LLM version.

- **`ENABLE_CUOTAS`** (default: `true`)
  - Controls:
    - Registration of `/add_cuotas` command
  - If `false`, cuotas columns still exist in DB, but users cannot create new installment expenses via command.

LLM configuration (only used if `ENABLE_LLM_FEATURES=true`):

- `GLM_API_KEY`
- `GLM_MODEL_NAME` (default `glm-4`)
- `GLM_BASE_URL` (default `https://open.bigmodel.cn/api/paas/v4`)
- `GLM_TEMPERATURE` (default `0.3`)
- `GLM_FALLBACK_TO_HF` (default `true`)
- `HUGGINGFACE_API_KEY`
- `HUGGINGFACE_MODEL` (default `mistralai/Mistral-7B-Instruct-v0.2`)

---

### 2. Code-level behavior of flags

- `ENABLE_LLM_FEATURES`:
  - If `false`:
    - No LLM client is created in `cmd/bot/main.go`.
    - `Handler.llmEnabled` is `false`.
    - `handleMessage` **never** calls `handleNaturalLanguageExpense`.
    - `/review` command is **not registered** on the router.
  - If `true` and `GLM_API_KEY` is set:
    - LLM client is initialized, `llmService` is plugged into:
      - Natural-language parsing for messages.
      - `AnalysisService.AnalyzeWithAI` used by `/review`.

- `ENABLE_CUOTAS`:
  - If `false`:
    - `/add_cuotas` is **not registered** on the router.
  - If `true`:
    - `/add_cuotas` is available and uses the new installment columns.

DB migrations for installments (in `internal/database/db.go`) are **additive only**:

- `ALTER TABLE expenses ADD COLUMN installment_group_id TEXT`
- `ALTER TABLE expenses ADD COLUMN installment_number INTEGER`
- `ALTER TABLE expenses ADD COLUMN installment_total INTEGER`
- `CREATE INDEX IF NOT EXISTS idx_expenses_installment_group ON expenses(installment_group_id)`

These do not break older binaries and do not need to be rolled back.

---

### 3. Preparing a Safe Rollout

1. **Tag current prod as rollback point**

   ```bash
   git checkout main
   git pull origin main
   git tag -a pre-llm-cuotas-$(date +%Y%m%d) -m "Pre LLM + cuotas rollout"
   git push origin pre-llm-cuotas-$(date +%Y%m%d)
   ```

   This tag is your one-command rollback reference.

2. **Deploy in “dark” mode first (flags off)**

   On your prod environment:

   ```bash
   export ENABLE_LLM_FEATURES=false
   export ENABLE_CUOTAS=false
   # keep existing TELEGRAM_BOT_TOKEN, DB_PATH, etc.
   ```

   Build and deploy the new commit (with migrations):

   ```bash
   go build ./cmd/bot
   # restart your service with the new binary
   ```

   This boots the new code + schema but without exposing new behavior.

3. **Smoke test with flags off**

   - `/add` works as before.
   - `/list`, `/summary`, `/settle`, etc. behave as before.
   - Natural-language messages **do nothing special** (no LLM parsing).
   - `/add_cuotas` is **not** available.

   If anything goes wrong at this step, use the rollback procedure below.

4. **Gradually enable features**

   a. **Enable cuotas only** (safer first step):

   ```bash
   export ENABLE_CUOTAS=true
   # leave ENABLE_LLM_FEATURES=false
   # restart bot process to pick up env change
   ```

   - Test `/add_cuotas` end-to-end in a test lobby.
   - Verify:
     - All installments created.
     - Dates/cycles look correct.
     - `/list` and `/list_billing` show the expected rows.

   b. **Enable LLM features** once cuotas look good:

   ```bash
   export ENABLE_LLM_FEATURES=true
   export GLM_API_KEY="..."
   # other GLM/HF env vars as needed
   # restart bot process
   ```

   - Test:
     - Sending free-text expense messages creates proper pending expenses.
     - Confirmation keyboard works and creates the correct expense(s).
     - `/review` runs and returns a reasonable AI summary (or a proper error if LLM is flaky).

---

### 4. Rollback Options

You have **three levels** of rollback, increasing in blast radius:

#### Level 1: Soft rollback via feature flags (preferred)

- To disable **LLM** immediately:

  ```bash
  export ENABLE_LLM_FEATURES=false
  # restart service
  ```

  Effects:
  - Natural-language parsing stops.
  - `/review` is no longer registered after restart.
  - Existing expenses remain; no schema change rollback needed.

- To disable **cuotas** immediately:

  ```bash
  export ENABLE_CUOTAS=false
  # restart service
  ```

  Effects:
  - `/add_cuotas` disappears.
  - Existing installment rows stay in the DB and still show up in lists.

#### Level 2: Code rollback to previous tag

If the bug is not easily isolated by flags (e.g. regression in shared logic), roll back to the tag:

```bash
git fetch origin
git checkout pre-llm-cuotas-YYYYMMDD
go build ./cmd/bot
# redeploy this binary and restart service
```

SQLite tolerates extra columns, so the old binary will keep working against the migrated DB.

#### Level 3: Infra-level rollback (containers, etc.)

If you build Docker images or have a release artifact per commit, keep:

- An image for `pre-llm-cuotas-YYYYMMDD`
- An image for the new rollout commit

Then rollback is just:

```bash
docker run ... your-image:pre-llm-cuotas-YYYYMMDD
```

or the equivalent in your orchestration tool.

---

### 5. Minimal Checklist Before Enabling Fully

- **Schema OK**
  - `expenses` table has `installment_group_id`, `installment_number`, `installment_total`.
  - Index `idx_expenses_installment_group` exists (non-fatal if missing, but nice to have).

- **Flags + env**
  - `ENABLE_LLM_FEATURES` and `ENABLE_CUOTAS` set explicitly in prod.
  - `GLM_API_KEY` present if `ENABLE_LLM_FEATURES=true`.

- **Functional smoke tests**
  - `/add` works (EN + ES users).
  - `/add_cuotas` works (if enabled).
  - Natural-language message creates a pending expense and confirmation works (if LLM enabled).
  - `/review` returns an answer or a clear error (if LLM enabled).

If any of these fail in prod, use **Level 1** rollback (flags) first; if that’s not enough, jump to **Level 2** with the git tag.

