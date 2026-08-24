# Codex human-gate: зачем он нужен и как выбрать права

Эта инструкция объясняет, как выбрать границу возможностей для Codex
human-gate run. Полный контракт CLI описан в [spec.md](spec.md).

В примерах предполагается, что `codex` уже выбран в project или user config.

## Зачем нужен human-gate

Обычный запуск `start-issue` передаёт управление интерактивной agent session.
Human-gate вместо этого запускает Codex в batch mode: агент может выполнить
issue без постоянного присутствия оператора в интерактивном UI, а в конце
обязан вернуть одно из двух явных состояний:

- `STATUS: DONE` — запрошенная работа завершена;
- `STATUS: HUMAN_GATE` — требуется реальное решение или недостающее разрешение,
  поэтому `start-issue` открывает оператору точный сохранённый Codex thread.

Такой режим нужен для полуавтономной работы над issue: реализация и тесты могут
выполняться без участия человека, но destructive actions, missing credentials
и product decisions по-прежнему останавливаются на видимом human gate.

Отдельные permission modes нужны потому, что изменение worktree и доставка PR
пересекают разные границы доверия. Codex в sandbox обычно может изменить и
проверить код, но может не получить доступ к GitHub, записи Git metadata, push
или созданию PR. Если сразу дать каждому run полный доступ, техническая проблема
исчезнет ценой незаметного расширения привилегий. Поэтому `start-issue`
сохраняет безопасную границу по умолчанию и требует явный opt-in, когда агент
должен самостоятельно доставить изменение до PR.

## Выбор режима

Для обычной реализации в worktree используйте `restricted`:

```bash
start-issue 123 --human-gate
```

`restricted` — режим по умолчанию. Codex работает в sandbox
`workspace-write`: он может изменять и тестировать файлы подготовленного
worktree, но network access, запись Git metadata, push и доставка PR не
гарантированы.

Если run должен также читать GitHub context, делать commit и push, создавать
или обновлять pull request, явно включите `full-delivery`:

```bash
start-issue 123 --human-gate \
  --human-gate-permissions full-delivery
```

Это явное согласие на запуск Codex без sandbox и approvals. Режим даёт
техническую возможность, но не разрешает destructive Git operations,
production/security changes или нерешённые product decisions. В таких случаях
Codex по-прежнему обязан вернуть `STATUS: HUMAN_GATE` и передать решение
оператору.

## Проверка перед full delivery

Проверьте выбранный GitHub account, remote и права на репозиторий:

```bash
gh auth status
git remote get-url origin
gh repo view --json nameWithOwner,viewerPermission
```

Сначала выполните dry-run: он покажет выбранный режим и launch command, но не
создаст worktree и не запустит Codex.

```bash
start-issue 123 --human-gate \
  --human-gate-permissions full-delivery \
  --dry-run
```

В output должна быть строка:

```text
Human-gate permissions: full-delivery (CLI)
```

## Пример: реализовать issue и доставить PR

Предположим, что issue `123` находится в текущем репозитории, а выбранный
GitHub account имеет write access.

1. Просмотрите план запуска:

   ```bash
   start-issue 123 --human-gate \
     --human-gate-permissions full-delivery \
     --dry-run
   ```

2. После проверки warning об unsandboxed execution запустите реальный run:

   ```bash
   start-issue 123 --human-gate \
     --human-gate-permissions full-delivery
   ```

3. Codex сможет реализовать и протестировать issue, сделать commit, push
   issue branch и создать или обновить pull request, если это допускают задача
   и состояние репозитория.

4. Финальный `STATUS: DONE` успешно завершит команду. Финальный
   `STATUS: HUMAN_GATE` откроет оператору точный сохранённый Codex thread.

CLI сохраняет диагностические данные в:

```text
<worktree>/.start-issue/runs/<timestamp>/events.jsonl
<worktree>/.start-issue/runs/<timestamp>/last-message.txt
<worktree>/.start-issue/runs/<timestamp>/thread-id
```

## Одноразовое переключение через environment

Environment variable удобно использовать для одной команды или в
контролируемой automation wrapper:

```bash
START_ISSUE_HUMAN_GATE_PERMISSIONS=full-delivery \
  start-issue 123 --human-gate
```

CLI option имеет приоритет над environment variable. Не стоит добавлять
`full-delivery` в глобальный shell profile: видимый opt-in рядом с командой
упрощает проверку границы unsandboxed execution.

## Диагностика

- Если `restricted` не может обратиться к GitHub или записать Git metadata,
  завершите доставку вручную либо повторите run с явно проверенным
  `full-delivery`.
- Если full delivery не может сделать push или создать PR, повторно проверьте
  `gh auth status`, выбранный account, `origin` и `viewerPermission`.
- При ошибке status parsing проверьте `events.jsonl` и `last-message.txt` в
  напечатанном state directory.
- Если automatic resume не сработал, используйте сохранённый thread id:

  ```bash
  codex resume --include-non-interactive <thread_id>
  ```

Встроенная справка:

```bash
start-issue --human-gate-help
```
