# Repository Instructions

## Memory Bank

This repository uses `memory-bank/` from `https://github.com/dapi/memory-bank/`.
Before non-trivial product, documentation, architecture, or feature work, read:

1. [memory-bank/README.md](memory-bank/README.md)
2. [memory-bank/product/context.md](memory-bank/product/context.md)
3. [memory-bank/domain/glossary.md](memory-bank/domain/glossary.md)
4. [memory-bank/engineering/testing-policy.md](memory-bank/engineering/testing-policy.md)
5. [memory-bank/ops/development.md](memory-bank/ops/development.md)

For new medium or large feature work, use
[memory-bank/flows/feature-flow.md](memory-bank/flows/feature-flow.md). New
feature packages should use `brief.md -> optional design.md ->
implementation-plan.md`. Existing legacy packages with `feature.md` and
`solution.md` may stay as-is until they are touched for real work.

## Local Checks

Run `make test` before handoff when changes affect code, tests, release logic,
or memory-bank navigation. It includes shell syntax checks, shellcheck,
memory-bank link audit, whitespace checks, and the Bats suite.
