# Спецификация: start-issue

## Обзор

**Название**: `start-issue`
**Тип**: Bash-скрипт
**Назначение**: автоматизировать начало работы над GitHub issue: получить issue через `gh`, создать git worktree, при необходимости запустить `init.sh`, переименовать zellij tab и запустить выбранный coding agent.

## Поддерживаемые агенты

Начальные значения `--agent`:

| Значение | Поведение |
|----------|-----------|
| `claude` | Запускает Claude Code. Это значение по умолчанию для обратной совместимости. |
| `codex` | Запускает Codex CLI в созданном worktree. |
| `kimi` | Запускает Kimi CLI в созданном worktree. |
| `pi` | Запускает Pi CLI из созданного worktree. |
| `none` | Готовит worktree и печатает ручные следующие шаги без запуска агента. |

## Входные данные

### Обязательные параметры

| Параметр | Формат | Примеры |
|----------|--------|---------|
| Issue | URL или номер | `https://github.com/owner/repo/issues/123` или `123` |

### Опциональные параметры

| Флаг | Описание | По умолчанию |
|------|----------|--------------|
| `--repo` / `-r` | Репозиторий `owner/repo` | Определяется из текущего `origin` remote |
| `--base` / `-b` | Базовая ветка | Из `origin/HEAD`, иначе текущая ветка |
| `--worktree-dir` / `-w` | Директория для worktree | `START_ISSUE_WORKTREE_DIR`, затем `~/worktrees` |
| `--flat` | Использовать плоский путь worktree, заменяя `/` на `-` | false |
| `--agent` | Агент: `claude`, `codex`, `kimi`, `pi`, `none` | См. приоритет выбора агента |
| `--no-agent` | Alias для `--agent none` | false |
| `--no-claude` | Совместимый alias для `--no-agent` | false |
| `--prompt` | Inline prompt template | См. приоритет prompt |
| `--prompt-file` | Файл prompt template | См. приоритет prompt |
| `--no-init` | Пропустить запуск `init.sh` | false |
| `--command` / `-c` | Совместимый Claude command для дефолтного Claude prompt | `/task-router:route-task` |
| `--ai` | Генерировать имя ветки выбранным агентом | false, используется быстрая bash-эвристика |
| `--dry-run` | Показать действия, не выполняя worktree/init/agent launch | false |

## Приоритет конфигурации

### Агент

1. CLI: `--agent`, `--no-agent`, `--no-claude`
2. Project config: `.start-issue/agent` в git top-level directory
3. User config: `~/.config/start-issue/agent`
4. Environment: `START_ISSUE_AGENT`
5. Built-in default: `claude`

Project root определяется через:

```bash
git rev-parse --show-toplevel
```

Если поддержка запуска вне git repo будет добавлена позже, fallback root может быть текущей директорией.

### Prompt

1. CLI: `--prompt-file path/to/prompt.md` или `--prompt "..."`
2. Project config: `.start-issue/prompt.md`
3. User config: `~/.config/start-issue/prompt.md`
4. Environment: `START_ISSUE_PROMPT_FILE` или `START_ISSUE_PROMPT`
5. Built-in default

Если одновременно заданы `--prompt-file` и `--prompt`, скрипт завершает работу с ошибкой. То же правило действует для `START_ISSUE_PROMPT_FILE` и `START_ISSUE_PROMPT`, когда env является активным источником prompt.

## Prompt templates

Claude без явного override использует совместимый plugin-native prompt:

```text
/task-router:route-task {ISSUE_URL}
```

Остальные агенты без явного override используют portable prompt:

```text
Implement GitHub issue {ISSUE_URL} in this worktree.

Context:
- Repo: {REPO}
- Issue: #{ISSUE_NUMBER}
- Title: {ISSUE_TITLE}
- Branch: {BRANCH_NAME}
- Worktree: {WORKTREE_PATH}

Start by reading the issue with gh if needed. Follow repository instructions. Keep changes scoped. Run relevant tests or checks. Summarize changed files and verification before finishing.
```

Поддерживаемые переменные:

| Переменная | Значение |
|------------|----------|
| `{ISSUE_URL}` | URL issue |
| `{ISSUE_NUMBER}` | Номер issue |
| `{ISSUE_TITLE}` | Заголовок issue |
| `{ISSUE_BODY}` | Body issue как plain text |
| `{ISSUE_LABELS}` | Labels через запятую |
| `{REPO}` | `owner/repo` |
| `{BRANCH_NAME}` | Имя созданной ветки |
| `{WORKTREE_PATH}` | Путь worktree |
| `{BASE_BRANCH}` | Базовая ветка |

Templating правила:

- `eval` не используется.
- Заменяются только известные placeholders.
- Неизвестные placeholders остаются без изменений.
- Multiline значения, включая `{ISSUE_BODY}`, вставляются как plain text.
- `--dry-run` печатает prompt source и launch command. Если rendered prompt очень большой, команда показывает placeholder, а полный prompt можно вывести через `START_ISSUE_DUMP_PROMPT=1`.

## Алгоритм работы

### Фаза 1: Валидация и парсинг

1. Распарсить CLI arguments.
2. Проверить зависимости: `git`, `gh`, `jq`, авторизацию `gh`.
3. Проверить, что текущая директория внутри git repo.
4. Определить project root через `git rev-parse --show-toplevel`.
5. Распарсить issue URL или issue number.
6. Определить repo из `origin` remote, если `--repo` не передан.
7. Определить base branch.
8. Выбрать agent по приоритету конфигурации.
9. Проверить наличие CLI выбранного agent, если agent не `none` и режим не `--dry-run`.

### Фаза 2: Получение issue

1. Получить данные issue через:

```bash
gh api "repos/{REPO}/issues/{ISSUE_NUMBER}"
```

2. Извлечь:

- title
- body
- labels
- issue URL

### Фаза 3: Имя ветки

По умолчанию используется быстрая bash-эвристика.

Правила типа ветки:

| Labels | Тип ветки |
|--------|-----------|
| `hotfix`, `critical`, `urgent` | `hotfix/` |
| `bug`, `fix`, `bugfix`, `error` | `fix/` |
| `docs`, `documentation` | `docs/` |
| `refactor`, `tech-debt`, `cleanup`, `technical` | `refactor/` |
| `test`, `testing`, `tests` | `test/` |
| `chore`, `ci`, `build`, `infra` | `chore/` |
| другое | `feature/` |

Формат:

```text
{type}/issue-{number}-{kebab-case-title}
```

`--ai` пытается сгенерировать имя ветки через выбранный agent в non-interactive mode и fallback-ится на bash-эвристику при ошибке или невалидном формате.

### Фаза 4: Создание worktree

1. Определить путь:

```text
{worktree-dir}/{branch-name}
```

Если включен `--flat`, `/` в имени ветки заменяется на `-`.

2. Если branch или worktree уже существуют, показать интерактивный выбор:

- использовать существующий worktree
- создать branch с suffix `-v2`, `-v3` и далее
- удалить и пересоздать
- выйти

3. Создать worktree:

```bash
git worktree add -b {BRANCH_NAME} {WORKTREE_PATH} origin/{BASE_BRANCH}
```

Если `origin/{BASE_BRANCH}` недоступен, используется `{BASE_BRANCH}`.

### Фаза 5: Инициализация окружения

Если `{WORKTREE_PATH}/init.sh` существует и не передан `--no-init`, выполнить:

```bash
cd {WORKTREE_PATH}
bash ./init.sh
```

Ненулевой exit code `init.sh` считается предупреждением, а не критической ошибкой.

### Фаза 6: Zellij

После успешного получения issue скрипт пытается выполнить соседний helper:

```bash
zellij-rename-tab-to-issue-number "{ISSUE_NUMBER}"
```

Если helper отсутствует, это не ошибка.

### Фаза 7: Запуск агента

Перед запуском выбирается prompt template, выполняется template substitution и формируется launch command.

Launch adapters:

```bash
claude:
  cd "$WORKTREE_PATH"
  exec claude --dangerously-skip-permissions "$PROMPT"

codex:
  exec codex --cd "$WORKTREE_PATH" --dangerously-bypass-approvals-and-sandbox "$PROMPT"

kimi:
  exec kimi --work-dir "$WORKTREE_PATH" --yolo -p "$PROMPT"

pi:
  cd "$WORKTREE_PATH"
  exec pi "$PROMPT"

none:
  print_manual_next_steps
```

Флаги проверены по установленным CLI help на 2026-04-21.

## Dry run

`--dry-run` не создает worktree, не запускает `init.sh` и не запускает agent. Он печатает:

- выбранный agent и источник
- worktree directory и источник
- выбранный prompt source
- длину rendered prompt
- команду запуска agent, которая была бы выполнена

## Обработка ошибок

Критические ошибки завершают скрипт с exit code 1:

| Ситуация | Сообщение |
|----------|-----------|
| Не в git repo | `Not in a git repository` |
| `gh` отсутствует | `gh CLI not found. Install: https://cli.github.com` |
| `gh` не авторизован | `gh not authenticated. Run: gh auth login` |
| `jq` отсутствует | `jq not found. Please install jq.` |
| Issue не найден | `Issue #{number} not found in {owner}/{repo}` |
| Agent неизвестен | `Unknown agent: {agent}` |
| Agent CLI отсутствует | `{agent} CLI not found. Install it or use --agent none.` |
| Prompt file отсутствует | `Prompt file not found: {path}` |
| Одновременно заданы inline и file prompt | `Use either ... not both.` |
| Worktree создать не удалось | `Failed to create worktree` |

Предупреждения не прерывают выполнение:

| Ситуация | Поведение |
|----------|-----------|
| `init.sh` отсутствует | Пропустить initialization |
| `init.sh` вернул ненулевой код | Напечатать warning и продолжить |
| AI branch naming не сработал | Использовать fast fallback |
| zellij helper отсутствует | Пропустить rename |

## Примеры использования

```bash
start-issue 123
start-issue https://github.com/owner/repo/issues/123
start-issue 123 --repo owner/repo
start-issue 123 --base develop
start-issue 123 --agent codex
start-issue 123 --agent kimi --prompt-file .start-issue/prompt.md
start-issue 123 --agent pi --prompt "Implement {ISSUE_URL} in {WORKTREE_PATH}"
start-issue 123 --no-agent
start-issue 123 --no-claude
START_ISSUE_AGENT=codex start-issue 123
START_ISSUE_WORKTREE_DIR=~/projects/worktrees start-issue 123
```

## Зависимости

Обязательные:

- `bash`
- `git`
- `gh` CLI с авторизованной GitHub session
- `jq`

Опциональные:

- `claude`, `codex`, `kimi`, `pi` - нужен только выбранный agent
- `init.sh` в корне worktree
- `zellij-rename-tab-to-issue-number` рядом со скриптом

## Критерии приемки

- [x] `start-issue 123` по умолчанию выбирает `claude`.
- [x] `start-issue 123 --agent codex` создает worktree и запускает Codex в этом worktree.
- [x] `start-issue 123 --agent kimi` запускает Kimi в этом worktree.
- [x] `start-issue 123 --agent pi` запускает Pi в этом worktree.
- [x] `start-issue 123 --no-agent` только готовит worktree и печатает следующие шаги.
- [x] Agent выбирается через CLI, `.start-issue/agent`, `~/.config/start-issue/agent`, `START_ISSUE_AGENT`.
- [x] Prompt выбирается через CLI, `.start-issue/prompt.md`, `~/.config/start-issue/prompt.md`, env.
- [x] Claude-specific aliases сохранены, help text описывает agent-neutral поведение.
- [x] `--dry-run` печатает selected agent, prompt source и launch command.
- [x] `START_ISSUE_WORKTREE_DIR` является env для worktree directory.
