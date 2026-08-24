#!/usr/bin/env python3
"""Render the README terminal demo from a real start-issue fixture run.

Requires Pillow. The GitHub and Codex boundaries are deterministic local fakes;
the start-issue binary and git worktree operations are real.
"""

from __future__ import annotations

import os
import re
import shlex
import shutil
import subprocess
import tempfile
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont


REPO_ROOT = Path(__file__).resolve().parents[1]
OUTPUT = REPO_ROOT / "docs" / "assets" / "start-issue-demo.gif"
FONT_CANDIDATES = (
    "/System/Library/Fonts/Menlo.ttc",
    "/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf",
)
FONT_SIZE = 16
WIDTH = 1100
HEIGHT = 680
LINE_HEIGHT = 23
TERMINAL_TOP = 58
TERMINAL_LEFT = 28
VISIBLE_LINES = 25

COLORS = {
    "background": "#0d1117",
    "chrome": "#161b22",
    "border": "#30363d",
    "text": "#e6edf3",
    "muted": "#8b949e",
    "green": "#3fb950",
    "blue": "#58a6ff",
    "cyan": "#39c5cf",
    "yellow": "#d29922",
    "purple": "#bc8cff",
}


def run(*args: str, cwd: Path | None = None, env: dict[str, str] | None = None) -> None:
    subprocess.run(args, cwd=cwd, env=env, check=True, stdout=subprocess.DEVNULL)


def write_executable(path: Path, content: str) -> None:
    path.write_text(content, encoding="utf-8")
    path.chmod(0o755)


def ensure_binary() -> Path:
    binary = REPO_ROOT / ".build" / "start-issue"
    if not binary.exists():
        subprocess.run(
            ["mise", "exec", "go@1.24", "--", "make", "build"],
            cwd=REPO_ROOT,
            check=True,
        )
    return binary


def capture_demo() -> list[str]:
    binary = ensure_binary()
    with tempfile.TemporaryDirectory(prefix="start-issue-readme-") as temporary:
        fixture = Path(temporary)
        repository = fixture / "repo"
        remote = fixture / "remote.git"
        home = fixture / "home"
        worktrees = home / "worktrees"
        fake_bin = fixture / "bin"

        for directory in (repository, worktrees, home / ".config" / "start-issue", fake_bin):
            directory.mkdir(parents=True, exist_ok=True)

        run("git", "init", "-q", "-b", "master", str(repository))
        run("git", "-C", str(repository), "config", "user.email", "demo@example.invalid")
        run("git", "-C", str(repository), "config", "user.name", "Demo User")
        (repository / "README.md").write_text("# Acme API\n", encoding="utf-8")
        run("git", "-C", str(repository), "add", "README.md")
        run("git", "-C", str(repository), "commit", "-q", "-m", "Initial commit")
        run("git", "init", "--bare", "-q", str(remote))
        run("git", "-C", str(repository), "remote", "add", "origin", str(remote))
        run("git", "-C", str(repository), "push", "-q", "-u", "origin", "master")
        (home / ".config" / "start-issue" / "agent").write_text("codex\n", encoding="utf-8")

        real_git = shutil.which("git")
        if real_git is None:
            raise SystemExit("git is required")
        write_executable(
            fake_bin / "git",
            f"""#!/bin/sh
if [ "$1 $2 $3" = "remote get-url origin" ]; then
  printf '%s\\n' 'git@github.com:acme/api.git'
  exit 0
fi
exec {shlex.quote(real_git)} "$@"
""",
        )

        write_executable(
            fake_bin / "gh",
            """#!/bin/sh
case "$1" in
  auth) exit 0 ;;
  api)
    printf '%s\\n' '{"number":42,"title":"Add OAuth login to dashboard","body":"Implement OAuth login flow.","labels":[{"name":"feature"},{"name":"auth"}]}'
    ;;
  *) exit 1 ;;
esac
""",
        )
        write_executable(
            fake_bin / "codex",
            """#!/bin/sh
printf '%s\\n' \\
  '╭──────────────────────────────────────────────────────────╮' \\
  '│ >_ OpenAI Codex                                         │' \\
  '│ directory: ~/worktrees/feature/issue-42-add-oauth…    │' \\
  '╰──────────────────────────────────────────────────────────╯' \\
  '' \\
  '› Implement issue #42: Add OAuth login to dashboard' \\
  '' \\
  '• I’ll inspect the repository and implement the issue.'
""",
        )

        environment = os.environ.copy()
        environment["HOME"] = str(home)
        environment["PATH"] = str(fake_bin) + os.pathsep + environment["PATH"]
        command = [
            str(binary),
            "42",
        ]
        result = subprocess.run(
            command,
            cwd=repository,
            env=environment,
            check=False,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
        )
        if result.returncode != 0:
            raise SystemExit(result.stdout)

        output = result.stdout
        output = output.replace(str(binary), "~/.local/bin/start-issue")
        output = output.replace(str(repository), "~/code/acme-api")
        output = output.replace(str(home), "~")
        output = re.sub(r"start-issue v[^\n]+", "start-issue v2.0.2", output)
        output = re.sub(r"HEAD is now at [0-9a-f]+", "HEAD is now at a18d47e", output)
        return output.rstrip().splitlines()


def line_color(line: str) -> str:
    stripped = line.strip()
    if stripped.startswith(("Handing off", "•")):
        return COLORS["green"]
    if stripped.startswith(("🔍", "›")):
        return COLORS["cyan"]
    if stripped.startswith("📁"):
        return COLORS["purple"]
    if stripped.startswith(("Preparing worktree", "branch '", "HEAD is now")):
        return COLORS["muted"]
    if stripped.startswith(("╭", "│", "╰")):
        return COLORS["blue"]
    if stripped.startswith("start-issue v"):
        return COLORS["yellow"]
    return COLORS["text"]


def draw_search_icon(draw: ImageDraw.ImageDraw, x: int, y: int) -> int:
    draw.ellipse((x + 2, y + 3, x + 13, y + 14), outline=COLORS["cyan"], width=2)
    draw.line((x + 12, y + 13, x + 18, y + 19), fill=COLORS["cyan"], width=2)
    return 24


def draw_folder_icon(draw: ImageDraw.ImageDraw, x: int, y: int) -> int:
    draw.rounded_rectangle((x + 1, y + 7, x + 19, y + 19), radius=2, fill=COLORS["purple"])
    draw.rectangle((x + 3, y + 4, x + 10, y + 8), fill=COLORS["purple"])
    return 25


def render_frame(lines: list[str], command: str, cursor: bool, font: ImageFont.FreeTypeFont) -> Image.Image:
    image = Image.new("RGB", (WIDTH, HEIGHT), COLORS["background"])
    draw = ImageDraw.Draw(image)
    draw.rounded_rectangle((1, 1, WIDTH - 2, HEIGHT - 2), radius=14, fill=COLORS["background"], outline=COLORS["border"], width=2)
    draw.rounded_rectangle((2, 2, WIDTH - 3, 44), radius=13, fill=COLORS["chrome"])
    draw.rectangle((2, 30, WIDTH - 3, 44), fill=COLORS["chrome"])
    for x, color in ((20, "#ff5f56"), (42, "#ffbd2e"), (64, "#27c93f")):
        draw.ellipse((x - 6, 16, x + 6, 28), fill=color)
    title = "~/code/acme-api — zsh"
    title_width = draw.textlength(title, font=font)
    draw.text(((WIDTH - title_width) / 2, 13), title, font=font, fill=COLORS["muted"])

    all_lines = ["$ " + command] + lines
    visible = all_lines[-VISIBLE_LINES:]
    y = TERMINAL_TOP
    for index, line in enumerate(visible):
        if index == 0 and len(all_lines) <= VISIBLE_LINES:
            draw.text((TERMINAL_LEFT, y), "$", font=font, fill=COLORS["green"])
            draw.text((TERMINAL_LEFT + 18, y), line[2:], font=font, fill=COLORS["text"])
            if cursor:
                cursor_x = TERMINAL_LEFT + 18 + draw.textlength(line[2:], font=font)
                draw.rectangle((cursor_x + 1, y + 2, cursor_x + 8, y + 18), fill=COLORS["text"])
        else:
            icon_offset = 0
            rendered = line
            if line.startswith("🔍"):
                icon_offset = draw_search_icon(draw, TERMINAL_LEFT, y)
                rendered = line[1:].lstrip()
            elif line.startswith("📁"):
                icon_offset = draw_folder_icon(draw, TERMINAL_LEFT, y)
                rendered = line[1:].lstrip()
            draw.text((TERMINAL_LEFT + icon_offset, y), rendered, font=font, fill=line_color(line))
        y += LINE_HEIGHT
    return image


def render_gif(output_lines: list[str]) -> None:
    font_path = next((path for path in FONT_CANDIDATES if Path(path).exists()), None)
    if font_path is None:
        raise SystemExit("No supported monospace font found")
    font = ImageFont.truetype(font_path, FONT_SIZE, index=0)
    command = "start-issue 42"
    frames: list[Image.Image] = []
    durations: list[int] = []

    frames.append(render_frame([], "", True, font))
    durations.append(500)
    for length in range(1, len(command) + 1, 2):
        frames.append(render_frame([], command[:length], True, font))
        durations.append(55)
    if len(command) % 2 == 0:
        frames.append(render_frame([], command, True, font))
        durations.append(250)

    revealed: list[str] = []
    for line in output_lines:
        revealed.append(line)
        frames.append(render_frame(revealed, command, False, font))
        if line.startswith(("🔍", "📁", "Handing off", "╭")):
            durations.append(500)
        elif not line:
            durations.append(220)
        else:
            durations.append(115)
    durations[-1] = 3200

    OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    palette = frames[-1].convert("P", palette=Image.Palette.ADAPTIVE, colors=96)
    paletted = [frame.quantize(palette=palette, dither=Image.Dither.NONE) for frame in frames]
    paletted[0].save(
        OUTPUT,
        save_all=True,
        append_images=paletted[1:],
        duration=durations,
        loop=0,
        optimize=True,
        disposal=1,
    )


def main() -> None:
    render_gif(capture_demo())
    size = OUTPUT.stat().st_size / 1024 / 1024
    print(f"Rendered {OUTPUT.relative_to(REPO_ROOT)} ({size:.2f} MiB)")


if __name__ == "__main__":
    main()
