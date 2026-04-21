# Примеры Агентов

## Claude

Claude является агентом по умолчанию, поэтому эти команды эквивалентны:

```bash
start-issue 123
start-issue 123 --agent claude
```

По умолчанию Claude получает repository-native команду task-router:

```text
/task-router:route-task {ISSUE_URL}
```

Используйте `--command`, чтобы сохранить стиль Claude slash-command, но заменить префикс команды:

```bash
start-issue 123 --agent claude --command "/debug"
```

Используйте project config, если Claude должен быть агентом по умолчанию для этого репозитория:

```bash
mkdir -p .start-issue
echo claude > .start-issue/agent
start-issue 123
```

## Codex

Запустить Codex для одного issue:

```bash
start-issue 123 --agent codex
```

Скрипт создает worktree, рендерит portable prompt и запускает:

```bash
codex --cd "$WORKTREE_PATH" --dangerously-bypass-approvals-and-sandbox "$PROMPT"
```

Использовать Codex как project default:

```bash
mkdir -p .start-issue
echo codex > .start-issue/agent
start-issue 123
```

Использовать custom prompt file с Codex:

```bash
start-issue 123 --agent codex --prompt-file .start-issue/prompt.md
```

## Kimi

Запустить Kimi для одного issue:

```bash
start-issue 123 --agent kimi
```

Скрипт создает worktree, рендерит portable prompt и запускает:

```bash
kimi --work-dir "$WORKTREE_PATH" --yolo -p "$PROMPT"
```

Использовать Kimi через environment, не меняя project files:

```bash
START_ISSUE_AGENT=kimi start-issue 123
```

Использовать inline prompt для одного запуска:

```bash
start-issue 123 --agent kimi --prompt "Implement {ISSUE_URL} in {WORKTREE_PATH}. Keep changes scoped and run tests."
```

## Pi

Запустить Pi для одного issue:

```bash
start-issue 123 --agent pi
```

Скрипт переходит в worktree и запускает:

```bash
cd "$WORKTREE_PATH"
pi "$PROMPT"
```

Использовать Pi как user default для всех репозиториев:

```bash
mkdir -p ~/.config/start-issue
echo pi > ~/.config/start-issue/agent
start-issue 123
```

Предварительно посмотреть, что именно будет выполнено перед запуском Pi:

```bash
start-issue 123 --agent pi --dry-run
```
