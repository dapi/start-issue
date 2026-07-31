# start-issue

[![CI](https://github.com/dapi/start-issue/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/dapi/start-issue/actions/workflows/ci.yml)

[English version](README.md)

Превращайте GitHub issue в отдельную ветку, git worktree и сессию coding agent.

`start-issue` превращает контекст issue в повторяемый workflow:

1. issue -> branch
2. branch -> worktree
3. worktree -> agent session

Он получает данные issue через `gh`, создает git worktree с именем ветки на основе issue, при необходимости запускает `init.sh`, опционально переименовывает текущую вкладку zellij и запускает настраиваемую сессию coding agent.

## Установка

Установить последний опубликованный релиз:

```bash
go install github.com/dapi/start-issue/v2/cmd/start-issue@latest
```

Опубликованные релизы содержат бинарники для конкретной платформы и файл
`checksums.txt`. После первоначальной установки через Go команда
`start-issue install` самостоятельно выбирает бинарник для текущей POSIX-платформы,
проверяет SHA-256 и устанавливает его в `~/.local/bin/start-issue`.

Ручная установка:

```bash
tmpdir="$(mktemp -d)"
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in x86_64) arch=amd64 ;; arm64|aarch64) arch=arm64 ;; esac
asset="start-issue-${os}-${arch}"
curl -fsSL "https://github.com/dapi/start-issue/releases/latest/download/${asset}" -o "$tmpdir/$asset"
curl -fsSL https://github.com/dapi/start-issue/releases/latest/download/checksums.txt -o "$tmpdir/checksums.txt"
(cd "$tmpdir" && grep -E " [*]?${asset}$" checksums.txt | shasum -a 256 -c -)
mkdir -p ~/.local/bin
install -m 0755 "$tmpdir/$asset" ~/.local/bin/start-issue
```

Сборка и установка из исходников:

```bash
make install
```

Команда собирает Go-бинарник `start-issue` из исходников и устанавливает его в `~/.local/bin/start-issue`.

Убедитесь, что `~/.local/bin` есть в вашем `PATH`.

Обновить уже установленный `start-issue` до последнего опубликованного GitHub Release:

```bash
start-issue update
start-issue --update
```

Workflow обновления определяет последний GitHub Release для `dapi/start-issue`,
сравнивает его с версией запущенного executable и, если доступен более новый
релиз, обновляет тот же путь executable. Если установленная версия уже актуальна,
команда успешно завершается и печатает понятный no-op статус.

## Codex Human-Gate

`start-issue 123 --agent codex --human-gate` сохраняет обычный issue-start workflow до создания worktree, опционального `init.sh` и рендера prompt, но заменяет финальный интерактивный запуск Codex на resumable batch run.

Batch flow:

1. запускает `codex exec` с JSON events и сохраненным `last-message` файлом;
2. получает `thread_id` из события `thread.started`;
3. завершает команду с кодом `0` на `STATUS: DONE`;
4. открывает `codex resume --include-non-interactive <thread_id>` на `STATUS: HUMAN_GATE`.

Режим намеренно поддерживается только для Codex. `--human-gate` с любым другим agent завершается явной ошибкой.

Отдельная справка:

```bash
start-issue --human-gate-help
```

Контракт prompt:

- Финальное сообщение должно содержать ровно одну terminal status line: `STATUS: DONE` или `STATUS: HUMAN_GATE`.
- `HUMAN_GATE` допустим только для реального пользовательского решения: destructive action, missing credentials, incompatible product choice или test failure, который нельзя безопасно исправить в scope issue.

Exit codes:

- `0`: Codex вернул `STATUS: DONE`.
- `1`: Codex завершился ошибкой, не был получен `thread_id`, не найден распознаваемый final status или failed parsing.
- `2`: Codex вернул `STATUS: HUMAN_GATE`, но `start-issue` не смог открыть interactive resume. Команда resume и `thread_id` печатаются для ручного запуска.

State files:

```text
<worktree>/.start-issue/runs/<timestamp>/events.jsonl
<worktree>/.start-issue/runs/<timestamp>/last-message.txt
<worktree>/.start-issue/runs/<timestamp>/thread-id
```

### Локальный E2E smoke test с реальным Codex

Обычный Bats-набор использует fake Codex CLI. Для проверки с реальным локальным
Codex из checkout `start-issue` выполните opt-in команду:

```bash
START_ISSUE_E2E=1 make e2e-human-gate
```

Скрипт использует приватный репозиторий `dapi/start-issue-e2e-fixture` и его
control issue, требует авторизованный `gh`, не допускает fake Codex и создаёт
отдельный временный clone и worktree parent. После успеха они удаляются; чтобы
сохранить их, задайте `START_ISSUE_E2E_KEEP=1`. Скрипт также отклоняет любые
изменения fixture worktree, кроме своего `.start-issue` state. Для проверки interactive resume:

```bash
START_ISSUE_E2E=1 \
test/e2e/human-gate.sh --scenario human-gate
```

Выйдите из возобновлённой Codex-сессии, после чего скрипт проверит артефакты.

#### Сценарии и проверки

| Сценарий | Команда | Что проверяется |
| --- | --- | --- |
| `done` | `START_ISSUE_E2E=1 make e2e-human-gate` | Реальный Codex batch run выдаёт `thread.started`, сохраняет `thread-id`, `events.jsonl` и `last-message.txt`, заканчивается `STATUS: DONE` и не меняет fixture worktree за пределами `.start-issue` state. |
| `human-gate` | `START_ISSUE_E2E=1 test/e2e/human-gate.sh --scenario human-gate` | Те же проверки артефактов и чистоты worktree, а также явный handoff `codex resume --include-non-interactive <thread_id>`. Перед завершением скрипта оператор выходит из возобновлённой interactive session. |

Оба сценария проверяют авторизованный `gh`, реальный, а не fake Codex binary, и
обязательный интерфейс справки `codex exec` (`--output-last-message`, без
устаревшего флага `--ask-for-approval`). Выбранный Codex executable печатается
в test output. Они не доказывают поведение приложения за пределами human-gate
protocol и намеренно не входят в CI.

## Использование

```bash
start-issue 123
start-issue https://github.com/owner/repo/issues/123
start-issue 123 --repo owner/repo --base develop
start-issue 123 --agent codex
start-issue 123 --agent codex --model gpt-5.2
start-issue 123 --agent codex --human-gate
start-issue 123 --agent kimi --prompt-file .start-issue/prompt.md
start-issue 123 --no-agent
start-issue 123 --dry-run
start-issue setup
start-issue --setup
start-issue init
start-issue init --project --agent codex --model gpt-5.2
start-issue update
start-issue --update
start-issue --human-gate-help
```

Запуск `start-issue` без issue печатает обычную справку, а также текущий
выбранный agent, выбранную model, источник prompt и расположение prompt, после
чего выходит без обращения к GitHub.

## Процесс

```mermaid
flowchart TD
    A["start-issue ISSUE [options]"] --> B["Определить контекст<br/>repo, issue, base branch"]
    B --> C["Загрузить конфигурацию<br/>agent, prompt, worktree dir"]
    C --> D["Получить metadata GitHub issue"]
    D --> Z["Опционально переименовать zellij tab<br/>через zellij-tab-status"]
    Z --> E["Спланировать branch<br/>и путь worktree"]
    E --> F{"--dry-run?"}

    F -- yes --> G["Напечатать план<br/>и выйти"]
    F -- no --> H["Создать или переиспользовать<br/>git worktree"]

    H --> I["Запустить hook инициализации<br/>если включен"]
    I --> J["Сформировать prompt<br/>для agent"]
    J --> K{"Agent выбран?"}

    K -- yes --> L["Запустить выбранный coding agent<br/>внутри worktree"]
    K -- no --> M["Напечатать ручные<br/>следующие шаги"]

    L --> N["Работа над issue"]
    M --> N
```

## Внутренняя архитектура

CLI entrypoint — `cmd/start-issue`; runtime, build и тесты реализованы на Go.
`make build` и `make install` собирают и устанавливают Go-бинарник.

- Helpers конфигурации и prompt разрешают CLI, environment, project и user defaults.
- Helpers repository/worktree получают metadata issue, безопасно планируют reuse и
  запускают optional hook `init.sh` внутри подготовленной worktree.
- Helpers agent adapter валидируют agent, строят launch commands, генерируют AI
  branch names и выполняют Codex human-gate mode.
- Helpers release выбирают platform assets, проверяют checksum и staged
  `--version`, затем атомарно устанавливают update.

Внутренний pipeline теперь такой:

1. Parse input.
2. Resolve config.
3. Fetch issue.
4. Plan branch and worktree.
5. Execute the plan.
6. Launch the selected agent.

Go implementation сохраняет lifecycle commands, configuration shape и output в
одном compiled CLI, оставляя `git`, `gh` и agent CLIs внешними process
boundaries. Новые возможности должны сохранять эти focused helper boundaries,
а не возвращать второй runtime.

## Аргументы CLI

| Аргумент | Описание |
|----------|----------|
| `ISSUE` | Номер GitHub issue или полный URL GitHub issue. Обязательный аргумент. |
| `init` | Создать файлы конфигурации по умолчанию для текущего проекта или текущего пользователя. |
| `setup` | Запустить first-run onboarding пользовательской конфигурации в `~/.config/start-issue`. |
| `update` | Обновить запущенный executable `start-issue` из последнего опубликованного GitHub Release. |
| `--repo OWNER/REPO`, `-r OWNER/REPO` | Репозиторий, из которого нужно прочитать issue, если `ISSUE` передан номером. Если не задан, `start-issue` определяет репозиторий из `origin`. |
| `--base BRANCH`, `-b BRANCH` | Базовая ветка для новой worktree branch. Если не задана, `start-issue` использует default branch репозитория, когда она доступна, иначе текущую ветку. |
| `--worktree-dir DIR`, `-w DIR` | Родительская директория для создаваемых worktree. Переопределяет `START_ISSUE_WORKTREE_DIR`. |
| `--flat` | Использовать плоский путь worktree, заменяя `/` в имени ветки на `-`. |
| `--agent AGENT` | Агент, который будет запущен после подготовки worktree. С `init` - agent по умолчанию, который нужно записать. Поддерживаются: `claude`, `codex`, `kimi`, `pi`, `none`. |
| `--model MODEL` | Явная model для выбранного агента. С `init` - model config, который нужно записать. Если не задана, встроенное поведение остается unset, и решение принимает CLI выбранного агента. |
| `--no-agent` | Подготовить worktree и напечатать ручные следующие шаги без запуска агента. Alias для `--agent none`. |
| `--no-claude` | Совместимый alias для `--no-agent`. |
| `--prompt TEXT` | Inline prompt template для выбранного агента. С `init` - prompt template, который нужно записать. Нельзя использовать вместе с `--prompt-file`. |
| `--prompt-file PATH` | Файл prompt template для выбранного агента. С `init` - содержимое файла, которое нужно записать. Нельзя использовать вместе с `--prompt`. |
| `--improve-prompt` | Попросить выбранного агента сгенерировать reviewable proposal улучшенного prompt template и выйти до создания worktree. |
| `--human-gate` | Codex-only batch mode для issue workflow. Запускает `codex exec`, выходит на `STATUS: DONE` и резюмирует ту же сессию на `STATUS: HUMAN_GATE`. |
| `--human-gate-help` | Показать отдельную справку по Codex human-gate workflow: prompt contract, exit codes и state files. |
| `--prompt-output-file PATH` | Путь для proposal-файла в режиме `--improve-prompt`. |
| `--no-init` | Не запускать `init.sh`, даже если он есть в созданном worktree. |
| `--command COMMAND`, `-c COMMAND` | Префикс Claude command для стандартного Claude prompt. Значение по умолчанию: `/task-router:route-task`. |
| `--ai` | Попросить выбранного агента сгенерировать имя ветки. Если генерация не удалась, используется локальная эвристика. |
| `--project` | С `init` - записать проектную конфигурацию в `.start-issue` в git root. |
| `--user` | С `init` - записать пользовательскую конфигурацию в `~/.config/start-issue`. |
| `--force` | С `init` - перезаписать существующие файлы `agent` и `prompt.md`, а также сбросить `model` к выбранному значению или к built-in unset, если `--model` не передан. Без `--force` существующие файлы сохраняются. |
| `--dry-run` | Напечатать выбранную конфигурацию и launch command без создания worktree, запуска `init.sh` или запуска агента. С `init` - напечатать план записи конфигурации без создания файлов. |
| `--setup` | Запустить тот же user-config onboarding flow, что и `start-issue setup`. |
| `--update` | Обновить запущенный executable `start-issue` из последнего опубликованного GitHub Release. Эквивалентно `start-issue update`. |
| `--version`, `-v` | Показать версию. |
| `--help`, `-h` | Показать справку. |

Подробные примеры по агентам находятся в [docs/agent-examples.ru.md](docs/agent-examples.ru.md).

Связанные Claude Code workflows из marketplace:

- [task-router](https://github.com/dapi/claude-code-marketplace/tree/master/task-router)
- [zellij-workflow](https://github.com/dapi/claude-code-marketplace/tree/master/zellij-workflow)

## Переменные окружения

| Переменная | Описание |
|------------|----------|
| `START_ISSUE_AGENT` | Агент по умолчанию, когда `--agent` не передан и agent не задан в файлах конфигурации. Поддерживаются: `claude`, `codex`, `kimi`, `pi`, `none`. Встроенное значение по умолчанию: `claude`. |
| `START_ISSUE_MODEL` | Model по умолчанию, когда `--model` не передан и model не задана в файлах конфигурации. Встроенное поведение по умолчанию: unset, решение принимает CLI выбранного агента. |
| `START_ISSUE_PROMPT` | Inline prompt template, который используется, если prompt не задан через CLI. Перебивает project и user prompt files. Нельзя использовать вместе с `START_ISSUE_PROMPT_FILE`, когда prompt не задан через CLI. |
| `START_ISSUE_PROMPT_FILE` | Файл prompt template, который используется, если prompt не задан через CLI. Перебивает project и user prompt files. Нельзя использовать вместе с `START_ISSUE_PROMPT`, когда prompt не задан через CLI. |
| `START_ISSUE_WORKTREE_DIR` | Родительская директория по умолчанию для создаваемых worktree, если `--worktree-dir` не передан. Встроенное значение по умолчанию: `~/worktrees`. |
| `START_ISSUE_DUMP_PROMPT` | Если задана в `1`, dry-run выводит полный rendered prompt вместо краткой информации. |

## Файлы конфигурации

| Файл | Описание |
|------|----------|
| `.start-issue/agent` | Агент по умолчанию для проекта. Читается из git root. |
| `.start-issue/model` | Model по умолчанию для проекта. Читается из git root, если файл существует. |
| `.start-issue/prompt.md` | Prompt template по умолчанию для проекта. Читается из git root. |
| `~/.config/start-issue/agent` | Пользовательский agent по умолчанию. |
| `~/.config/start-issue/model` | Пользовательская model по умолчанию. Читается, если файл существует. |
| `~/.config/start-issue/prompt.md` | Пользовательский prompt template по умолчанию. |

Для дружелюбного user-level onboarding используйте `start-issue setup` или `start-issue --setup`. Команда работает только с `~/.config/start-issue`, спрашивает default agent (`claude`, `codex`, `kimi`, `pi` или skip), показывает derived default prompt и сохраняет `prompt.md` только после явного подтверждения.

Для существующего manual initializer используйте `start-issue init`. Если не переданы `--project` или `--user`, команда спросит, какой scope инициализировать. Она записывает встроенные agent и prompt по умолчанию, если не переданы `--agent`, `--prompt` или `--prompt-file`. `--model` записывает соседний файл `model`; если `--model` не передан, встроенное поведение остается unset и новый model-файл не создается. Если существующий файл `agent` сохраняется без `--force`, default prompt выбирается для этого сохраненного agent.

При обычном non-setup запуске, если `~/.config/start-issue` еще не существует, `start-issue` показывает компактное first-run сообщение и спрашивает, нужно ли сразу запустить setup. Если пользователь отказывается, команда все равно создает пустую директорию `~/.config/start-issue`, чтобы onboarding автоматически не повторялся.

## Self-update

`start-issue update` и `start-issue --update` являются эквивалентными entry point.

Workflow:

1. Определяет последний опубликованный GitHub Release для `dapi/start-issue`.
2. Читает версию executable, который пользователь запустил.
3. Нормализует версии, поэтому `1.11.1` и `v1.11.1` считаются равными.
4. Если текущая версия уже актуальна или новее последнего опубликованного релиза, команда завершается с кодом `0` и печатает понятный статус.
5. Если опубликован более новый релиз, команда скачивает бинарник для текущей платформы и `checksums.txt`, проверяет checksum и устанавливает обновление в resolved target executable path, который был вызван.

Workflow обновления работает вне git repository и требует только `gh`. Go-бинарник
сам разбирает metadata release, скачивает assets и проверяет checksums.

Приоритет конфигурации:

1. Agent: CLI `--agent` / `--no-agent`, затем project config, user config, `START_ISSUE_AGENT`, затем built-in default `claude`
2. Model: CLI `--model`, затем project config, user config, `START_ISSUE_MODEL`, затем built-in unset
3. Prompt: CLI, затем project config, user config, env prompt, затем built-in default

Claude по умолчанию использует plugin-native команду:

```text
/task-router:route-task {ISSUE_URL}
```

Другие агенты по умолчанию используют portable prompt.

Чтобы улучшить prompt template, который будет использоваться для будущих стартов разработки, запустите:

```bash
start-issue 123 --agent codex --improve-prompt
```

Команда выбирает активный prompt template по обычному приоритету, получает issue как контекст, просит выбранного агента вернуть полный улучшенный prompt template и записывает proposal-файл. Активный prompt не перезаписывается. Для Markdown prompt-файлов proposal по умолчанию создается рядом с источником как `*.improved.md`; к остальным именам файлов добавляется `.improved`. Для built-in и inline prompt используется `.start-issue/prompt.improved.md`. Используйте `--prompt-output-file`, чтобы указать другой путь.

Prompt templates поддерживают:

```text
{ISSUE_URL}
{ISSUE_NUMBER}
{ISSUE_TITLE}
{ISSUE_BODY}
{ISSUE_LABELS}
{REPO}
{BRANCH_NAME}
{WORKTREE_PATH}
{BASE_BRANCH}
```

Неизвестные placeholders остаются без изменений.

## Поддержка Zellij

Если [`zellij-tab-status`](https://github.com/dapi/zellij-tab-status) доступен в `PATH`, `start-issue` после получения issue переименовывает текущую вкладку Zellij в `#ISSUE_NUMBER` через `zellij-tab-status --set-name`.

Этот шаг опциональный. Отсутствие `zellij-tab-status` игнорируется, а ошибка переименования выводится как предупреждение и не останавливает workflow.

Опциональная зависимость для поддержки Zellij:

- [`zellij-tab-status`](https://github.com/dapi/zellij-tab-status)

## Требования

- `git`
- `gh` CLI с авторизованной GitHub session
- CLI выбранного агента, если не используется `--agent none` или `--dry-run`

Для сборки из исходников дополнительно нужен Go 1.24+. Опциональному Bash
installer и manual-install snippet нужны `bash`, `curl` или `wget` и SHA-256
tool; `start-issue update` эти инструменты не использует.

## Релизы

GitHub Releases публикуются автоматически, когда в репозиторий пушится SemVer tag вроде `v2.0.0`. Release workflow заново прогоняет тесты, собирает Go-бинарники для поддерживаемых платформ и загружает:

- `start-issue-<os>-<arch>` (и `.exe` для Windows)
- `checksums.txt`
- `start-issue` и `start-issue.sha256` (временный bridge для обновления с v1)

Подготовить релиз локально можно так:

```bash
make test
git tag v2.0.0
git push origin v2.0.0
```

Перед подготовкой релиза добавьте user-facing изменения в `CHANGELOG.md` под `## [Unreleased]`.

Создавайте релизы из чистого рабочего дерева после успешных `make test` и `make build`. Tag определяет версию публикуемого бинарного файла.

Опубликовать подготовленный release можно также вместе с веткой:

```bash
git push origin master --follow-tags
```

## Спецификация

Спецификация скрипта находится в [doc/spec.md](doc/spec.md).

## Лицензия

MIT
