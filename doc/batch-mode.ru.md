# Batch mode: автономная работа Codex с остановкой для человека

Эта инструкция написана для человека, который впервые видит флаг
`start-issue --batch` и хочет понять, что он делает. Здесь объясняется, чем
batch mode отличается от обычного запуска, зачем Codex может остановиться со
статусом `HUMAN_GATE`, как после этого продолжить ту же session и какие права
нужны агенту для работы с кодом или полной доставки pull request.

В примерах предполагается, что `codex` уже выбран в project или user config.
Ранее выпущенный флаг `--human-gate` продолжает работать как совместимый alias
для `--batch`. Полное техническое описание находится в [spec.md](spec.md).

## Что такое human-gate

Human-gate — это контрольная точка, в которой автономная работа Codex либо
считается завершённой, либо передаётся человеку для принятия решения.

При обычном запуске `start-issue` открывает интерактивную Codex session, и
пользователь остаётся внутри неё. С флагом `--batch` Codex запускается в
batch mode и самостоятельно работает над issue. Пользователю не нужно следить
за каждым шагом в интерактивном UI.

В конце Codex обязан явно сообщить один из двух результатов:

- `STATUS: DONE` — задача выполнена и вмешательство человека не требуется;
- `STATUS: HUMAN_GATE` — продолжать без решения человека нельзя.

Human-gate — это не запрос подтверждения перед каждой командой и не название
sandbox. Это договорённость о том, как автономный запуск завершается и когда
управление возвращается человеку.

## Зачем он нужен

Без human-gate трудно надёжно отличить завершённую задачу от запуска, который
остановился из-за вопроса, ошибки или нехватки доступа. Одного exit code Codex
для этого недостаточно.

Human-gate позволяет:

- отдать Codex реализацию issue и выполнение тестов без постоянного наблюдения;
- получить однозначный результат `DONE`, когда работа действительно закончена;
- остановиться на решении, которое агент не должен принимать самостоятельно;
- продолжить тот же Codex thread со всей историей, а не объяснять задачу новому
  агенту заново.

## Как он работает

1. `start-issue` получает issue, создаёт или переиспользует worktree и готовит
   prompt так же, как при обычном запуске.
2. Вместо интерактивной session запускается `codex exec` в batch mode.
3. `start-issue` сохраняет события запуска, финальное сообщение и `thread_id`.
4. Codex выполняет задачу и заканчивает ответ ровно одной status line:
   `STATUS: DONE` или `STATUS: HUMAN_GATE`.
5. При `DONE` команда завершается с кодом `0`.
6. При `HUMAN_GATE` команда открывает сохранённую session через
   `codex resume --include-non-interactive <thread_id>`. Пользователь видит всю
   историю, отвечает на вопрос и продолжает работу в том же thread.
7. Если Codex завершился с ошибкой или не вернул понятный status,
   `start-issue` сообщает об ошибке и оставляет диагностические файлы.

Human-gate срабатывает не по ключевым словам в issue и не после каждой команды.
Решение остановиться принимает Codex по правилам prompt, а `start-issue` читает
status line из его финального сообщения и выполняет нужный переход.

## Когда Codex должен остановиться

`STATUS: HUMAN_GATE` нужен, когда без человека нельзя безопасно выбрать
следующий шаг. Например:

- требуется destructive operation, удаление данных или переписывание Git
  history;
- отсутствуют credentials или доступ, который должен предоставить человек;
- есть несколько несовместимых product-вариантов без зафиксированного решения;
- тесты или конфликт нельзя безопасно исправить в рамках issue;
- действие затрагивает production или security и требует отдельного согласия.

Обычные технические вопросы, исправимые ошибки и выбор реализации внутри
согласованного scope не должны останавливать run: Codex должен решить их сам.

## Как permission mode связан с human-gate

Это две разные настройки:

- `--batch` включает автономный batch run и протокол
  `DONE` / `HUMAN_GATE`;
- `--batch-permissions` определяет, какие технические действия доступны
  Codex во время этого run.

Разделение нужно потому, что правка кода и доставка PR требуют разных прав.
Sandbox обычно позволяет менять и тестировать файлы worktree, но может не дать
доступ к GitHub, записи Git metadata, push или созданию PR. Поэтому безопасный
режим остаётся default, а полный доступ для end-to-end доставки включается
явно.

## Выбор режима

Для обычной реализации в worktree используйте `restricted`:

```bash
start-issue 123 --batch
```

`restricted` — режим по умолчанию. Codex работает в sandbox
`workspace-write`: он может изменять и тестировать файлы подготовленного
worktree, но network access, запись Git metadata, push и доставка PR не
гарантированы.

Если run должен также читать GitHub context, делать commit и push, создавать
или обновлять pull request, явно включите `full-delivery`:

```bash
start-issue 123 --batch \
  --batch-permissions full-delivery
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
start-issue 123 --batch \
  --batch-permissions full-delivery \
  --dry-run
```

В output должна быть строка:

```text
Batch permissions: full-delivery (CLI)
```

## Пример: реализовать issue и доставить PR

Предположим, что issue `123` находится в текущем репозитории, а выбранный
GitHub account имеет write access.

1. Просмотрите план запуска:

   ```bash
   start-issue 123 --batch \
     --batch-permissions full-delivery \
     --dry-run
   ```

2. После проверки warning об unsandboxed execution запустите реальный run:

   ```bash
   start-issue 123 --batch \
     --batch-permissions full-delivery
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
START_ISSUE_BATCH_PERMISSIONS=full-delivery \
  start-issue 123 --batch
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
start-issue --batch-help
```
