---
title: Feature Packages Index
doc_kind: feature
doc_function: index
purpose: Навигация по instantiated feature packages. Читать, чтобы найти существующую delivery-единицу или понять, где создавать новую.
derived_from:
  - ../dna/governance.md
  - ../flows/feature-flow.md
status: active
audience: humans_and_agents
---

# Feature Packages Index

Каталог `memory-bank/features/` хранит instantiated feature packages вида `FT-XXX/`.

## Rules

- Каждый package создается по правилам из [`../flows/feature-flow.md`](../flows/feature-flow.md).
- Bootstrap package начинается с `README.md` и `brief.md`; после `Problem Ready` в него добавляется `design.md`, если `brief.md` фиксирует `Design required: yes`; `implementation-plan.md` появляется после готовности нужных upstream owners.
- Для bootstrap и downstream-документов используй шаблоны из [`../flows/templates/feature/`](../flows/templates/feature/).
- Если работа требует roadmap, risk register и нескольких delivery subissues, сначала создай или обнови epic package в [`../epics/README.md`](../epics/README.md).
- По умолчанию feature ссылается на общий product context из [`../product/context.md`](../product/context.md), а при изменении предметных правил также на соответствующие документы из [`../domain/README.md`](../domain/README.md).
- Если feature реализует или существенно меняет устойчивый сценарий проекта, она должна ссылаться на соответствующий `UC-*` из [`../use-cases/README.md`](../use-cases/README.md).
- В шаблонном репозитории этот каталог может быть пустым. Это нормально.

## Naming

- Базовый формат: `FT-XXX/`
- Вместо `XXX` используй идентификатор, принятый в проекте: issue id, ticket id или другой стабильный ключ
- Один package = одна delivery-единица

## Feature packages

- [Missing feature packages inventory](missing.md)
  Comparison of GitHub issues with instantiated packages, including current
  gaps, duplicates, and superseded migration work.

- [FT-016: Real Codex human-gate E2E suite](FT-016/README.md)
  Opt-in local smoke suite that validates the real Codex human-gate path.

- [FT-017: Codex human-gate delivery permissions](FT-017/README.md)
  Explicit restricted/default and opt-in full-delivery capability contract for
  Codex human-gate runs.
