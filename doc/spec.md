# Спецификация: start-issue

## Обзор

**Название**: `start-issue`
**Тип**: Bash CLI с модульной shell-архитектурой
**Назначение**: автоматизировать начало работы над GitHub issue: получить issue через `gh`, опционально переименовать zellij tab через `zellij-tab-status`, создать git worktree, при необходимости запустить `init.sh` и запустить выбранный coding agent.

Для разработки код остается модульным в `scripts/lib/start_issue/`, но install/distribution path должен собирать self-contained single-file script.

## Поддерживаемые агенты

Начальные значения `--agent`:

| Значение | Поведение |
|----------|-----------|
| `claude` | Запускает Claude Code. Это значение по умолчанию для обратной совместимости. |
| `codex` | Запускает Codex CLI в созданном worktree. |
| `kimi` | Запускает Kimi CLI в созданном worktree. |
| `pi` | Запускает Pi CLI из созданного worktree. |
| `none` | Готовит worktree и печатает ручные следующие шаги без запуска агента. |

## Внутренняя архитектура

Публичный CLI contract остается за `scripts/start-issue`, но внутренняя реализация разбита на shell-модули в `scripts/lib/start_issue/`.

| Модуль | Ответственность |
|--------|-----------------|
| `cli.sh` | CLI parsing и нормализация входных флагов |
| `config.sh` | Разрешение agent/prompt config и prompt rendering |
| `github.sh` | Parse issue input, detect repo/base branch, fetch issue metadata |
| `worktree.sh` | Branch naming, worktree planning, worktree/init/zellij side effects |
| `agent.sh` | Agent adapter contract: validate, branch-name generation, prompt improvement, launch command |
| `output.sh` | Help, status rendering, dry-run output, session header |
| `init.sh` | Workflow конфигурационного `init` |
| `pipeline.sh` | Явная orchestration pipeline |

Agent-specific behavior должен быть централизован за единым adapter boundary:

- validate agent support
- build launch command
- generate branch name in `--ai`
- improve prompt template in `--improve-prompt`

Если будущие изменения потребуют nested configuration, richer lifecycle subcommands (`resume`, `list`, `cleanup`) или полноценный structured output, это считается порогом для оценки Python core вместо дальнейшего роста Bash.

## Входные данные

### Обязательные параметры

| Параметр | Формат | Примеры |
|----------|--------|---------|
| Issue | URL или номер | `https://github.com/owner/repo/issues/123` или `123` |
| Config init | Литерал `init` | `start-issue init` |

Если Issue не передан, команда печатает справку и текущую конфигурацию:
выбранный agent с источником, prompt source, prompt location и короткий prompt
preview. Этот путь не обращается к GitHub и завершается с ненулевым кодом.

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
| `--improve-prompt` | Сгенерировать reviewable proposal улучшенного prompt template и выйти до создания worktree | false |
| `--prompt-output-file` | Путь proposal-файла для `--improve-prompt` | Для prompt-файла: рядом с source как `*.improved.md`; иначе `.start-issue/prompt.improved.md` |
| `--no-init` | Пропустить запуск `init.sh` | false |
| `--command` / `-c` | Совместимый Claude command для дефолтного Claude prompt | `/task-router:route-task` |
| `--ai` | Генерировать имя ветки выбранным агентом | false, используется быстрая bash-эвристика |
| `--project` | Для `init`: записать конфигурацию проекта в `.start-issue` | интерактивный выбор |
| `--user` | Для `init`: записать пользовательскую конфигурацию в `~/.config/start-issue` | интерактивный выбор |
| `--force` | Для `init`: перезаписать существующие `agent` и `prompt.md` | false |
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
2. Environment: `START_ISSUE_PROMPT_FILE` или `START_ISSUE_PROMPT`
3. Project config: `.start-issue/prompt.md`
4. User config: `~/.config/start-issue/prompt.md`
5. Built-in default

Если одновременно заданы `--prompt-file` и `--prompt`, скрипт завершает работу с ошибкой. То же правило действует для `START_ISSUE_PROMPT_FILE` и `START_ISSUE_PROMPT`, когда prompt не задан через CLI.

### Prompt improvement

`--improve-prompt` включает режим улучшения prompt template, который используется для старта разработки.

Алгоритм режима:

1. Выбрать active prompt template по обычному приоритету.
2. Получить GitHub issue, чтобы использовать его как контекст улучшения.
3. Попросить выбранного agent вернуть полный улучшенный prompt template.
4. Записать результат в reviewable proposal-файл.
5. Завершить выполнение до переименования Zellij tab, генерации branch, создания worktree, запуска `init.sh` и запуска agent session.

Режим не перезаписывает active prompt template. Если active prompt взят из файла, proposal по умолчанию пишется рядом с ним как `*.improved.md`. Если active prompt built-in или inline, proposal по умолчанию пишется в `.start-issue/prompt.improved.md` в git top-level directory. `--prompt-output-file` задает путь явно.

Если proposal-файл уже существует, скрипт завершается с ошибкой, чтобы не перезаписать reviewable артефакт. `--agent none` в этом режиме невалиден.

## Инициализация конфигурации

`start-issue init` создает файлы конфигурации:

- project scope: `{git-root}/.start-issue/agent` и `{git-root}/.start-issue/prompt.md`
- user scope: `~/.config/start-issue/agent` и `~/.config/start-issue/prompt.md`

Если не передан `--project` или `--user`, команда интерактивно спрашивает scope. Project scope требует запуск внутри git repository; user scope может выполняться вне git repository. Режим `init` не требует issue, `gh` или `jq`.

По умолчанию записывается agent `claude` и стандартный Claude prompt. `--agent` меняет записываемый agent; `--prompt` или `--prompt-file` меняют записываемый prompt. Если выбран не `claude` и prompt явно не задан, записывается portable prompt. Если существующий `agent` сохраняется без `--force`, default prompt выбирается по сохраненному agent, а не по built-in default или CLI override.

Существующие файлы не перезаписываются. `--force` перезаписывает `agent` и `prompt.md`.

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

Оркестрация выражена явным pipeline:

1. Parse input.
2. Resolve config.
3. Fetch issue.
4. Plan branch and worktree.
5. Execute the plan.
6. Launch the selected agent.

### Фаза 0: Config init

Если первый positional argument равен `init`:

1. Прочитать `--project`, `--user`, `--force`, `--agent`, `--prompt`, `--prompt-file`, `--command`.
2. Если scope не задан, спросить пользователя: project config или user config.
3. Для project config проверить git repository и определить git root.
4. Выбрать фактический agent: существующий `agent`, если он сохраняется без `--force`; иначе `--agent`; иначе `claude`.
5. Выбрать prompt: `--prompt-file`, `--prompt`, иначе built-in prompt для выбранного agent.
6. Создать target directory.
7. Записать `agent` и `prompt.md`; существующие файлы оставить без изменений, если не передан `--force`. В `--dry-run` только напечатать planned writes.
8. Завершить работу без получения issue, создания worktree и запуска agent.

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
10. Выбрать prompt template.
11. Если включен `--improve-prompt`, сгенерировать proposal улучшенного prompt template и завершить workflow до worktree/agent launch.

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

### Фаза 3: Zellij

После успешного получения issue скрипт проверяет наличие `zellij-tab-status` в `PATH`.

Если `zellij-tab-status` установлен и режим не `--dry-run`, выполняется:

```bash
zellij-tab-status --set-name "#{ISSUE_NUMBER}"
```

Если `zellij-tab-status` отсутствует, это не ошибка. Если команда переименования завершилась с ошибкой, скрипт печатает warning и продолжает workflow.

В режиме `--dry-run` скрипт печатает, был бы выполнен rename или шаг был бы пропущен из-за отсутствия `zellij-tab-status`.

### Фаза 4: Имя ветки

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

### Фаза 5: Создание worktree

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

### Фаза 6: Инициализация окружения

Если `{WORKTREE_PATH}/init.sh` существует и не передан `--no-init`, выполнить:

```bash
cd {WORKTREE_PATH}
bash ./init.sh
```

Ненулевой exit code `init.sh` считается предупреждением, а не критической ошибкой.

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
- planned/skip информацию для optional zellij tab rename
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
| Agent CLI отсутствует вне `--dry-run` | `{agent} CLI not found. Install it or use --agent none.` |
| Prompt file отсутствует | `Prompt file not found: {path}` |
| `--improve-prompt` используется с `--agent none` | `--improve-prompt requires an agent. Use --agent claude, codex, kimi, or pi.` |
| Proposal-файл уже существует | `Prompt improvement output already exists: {path}` |
| Agent не смог сгенерировать proposal | `Could not generate improved prompt with {agent}` |
| Одновременно заданы inline и file prompt | `Use either ... not both.` |
| Worktree создать не удалось | `Failed to create worktree` |
| Issue не передан | Печатает help и current configuration, затем выходит с ненулевым кодом |

Предупреждения не прерывают выполнение:

| Ситуация | Поведение |
|----------|-----------|
| `init.sh` отсутствует | Пропустить initialization |
| `init.sh` вернул ненулевой код | Напечатать warning и продолжить |
| AI branch naming не сработал | Использовать fast fallback |
| `zellij-tab-status` отсутствует | Пропустить rename |
| `zellij-tab-status --set-name` вернул ненулевой код | Напечатать warning и продолжить |

## Примеры использования

```bash
start-issue 123
start-issue https://github.com/owner/repo/issues/123
start-issue 123 --repo owner/repo
start-issue 123 --base develop
start-issue 123 --agent codex
start-issue 123 --agent kimi --prompt-file .start-issue/prompt.md
start-issue 123 --agent codex --improve-prompt
start-issue 123 --agent codex --improve-prompt --prompt-output-file .start-issue/prompt.next.md
start-issue 123 --agent pi --prompt "Implement {ISSUE_URL} in {WORKTREE_PATH}"
start-issue 123 --no-agent
start-issue 123 --no-claude
start-issue init
start-issue init --project --agent codex
start-issue init --user --force
START_ISSUE_AGENT=codex start-issue 123
START_ISSUE_WORKTREE_DIR=~/projects/worktrees start-issue 123
```

## Зависимости

Обязательные:

- `bash`
- `git`
- `gh` CLI с авторизованной GitHub session
- `jq`

Для `start-issue init --user` обязателен только `bash`. Для `start-issue init --project` нужны `bash` и `git`.

Опциональные:

- `claude`, `codex`, `kimi`, `pi` - нужен только выбранный agent, если не используется `--dry-run`
- `init.sh` в корне worktree
- `zellij-tab-status` в `PATH` для поддержки переименования вкладки Zellij

## Критерии приемки

- [x] `start-issue 123` по умолчанию выбирает `claude`.
- [x] `start-issue 123 --agent codex` создает worktree и запускает Codex в этом worktree.
- [x] `start-issue 123 --agent kimi` запускает Kimi в этом worktree.
- [x] `start-issue 123 --agent pi` запускает Pi в этом worktree.
- [x] `start-issue 123 --no-agent` только готовит worktree и печатает следующие шаги.
- [x] Agent выбирается через CLI, `.start-issue/agent`, `~/.config/start-issue/agent`, `START_ISSUE_AGENT`.
- [x] Prompt выбирается через CLI, `.start-issue/prompt.md`, `~/.config/start-issue/prompt.md`, env.
- [x] Запуск без Issue печатает выбранный agent и prompt details с расположением prompt.
- [x] `start-issue init` создает project или user config с agent и prompt по умолчанию.
- [x] `start-issue init --force` перезаписывает существующие config-файлы.
- [x] `--improve-prompt` создает reviewable proposal улучшенного prompt template и не перезаписывает active prompt.
- [x] Claude-specific aliases сохранены, help text описывает agent-neutral поведение.
- [x] `--dry-run` печатает selected agent, prompt source и launch command.
- [x] `START_ISSUE_WORKTREE_DIR` является env для worktree directory.
- [x] Если `zellij-tab-status` установлен, `start-issue` опционально переименовывает вкладку Zellij.
