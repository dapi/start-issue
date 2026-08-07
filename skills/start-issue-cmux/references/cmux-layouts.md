# cmux layout patterns

Start with a single terminal surface. Add only the surfaces needed by the issue.

## Test/log helper pane

```bash
cmux new-pane --workspace workspace:16 --type terminal --direction right --focus false
```

Send commands only to the returned surface, and only after the user has named the target workspace or requested this layout.

## Browser verification

Use the official `cmux-browser` skill for browser actions. Keep the browser in the same named workspace as the agent and do not copy authentication data into prompts or workspace metadata.

## Markdown handoff

Use the official `cmux-markdown` skill to show an existing plan or handoff file. Do not create a second live state journal in cmux; repository state, issue state, and agent state remain canonical.

## Project customization

Use `cmux-customization` for reusable `.cmux/` project layouts, actions, buttons, and Dock controls. Keep project-specific commands in the project config; keep this skill generic.
