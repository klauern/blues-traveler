#!/usr/bin/env python3

from __future__ import annotations

import pathlib
import re
import sys


ROOT = pathlib.Path(__file__).resolve().parent.parent
SCOPES = [
    ROOT / "docs" / "hooks-corpus",
    ROOT / "research-notes",
]
EXTRA_FILES = [
    ROOT / "docs" / "development" / "beads-workflow.md",
]
LINK_RE = re.compile(r"\[[^][]+\]\(([^)]+)\)")
ABS_PATH_RE = re.compile(r"/Users/klauer(?:/|$)")


def iter_files() -> list[pathlib.Path]:
    files: list[pathlib.Path] = []
    for scope in SCOPES:
        files.extend(sorted(scope.rglob("*.md")))
    files.extend(EXTRA_FILES)
    return files


def main() -> int:
    failures = 0
    for file_path in iter_files():
        rel_file = file_path.relative_to(ROOT)
        text = file_path.read_text(encoding="utf-8")

        if ABS_PATH_RE.search(text):
            print(f"absolute-path: {rel_file}")
            failures = 1

        if rel_file == pathlib.Path("docs/hooks-corpus/templates/system-template.md"):
            continue

        for match in LINK_RE.finditer(text):
            target = match.group(1)
            if target.startswith(("http://", "https://", "mailto:", "#")):
                continue

            target = target.split("#", 1)[0]
            if not target:
                continue

            resolved = (file_path.parent / target).resolve()
            if not resolved.exists():
                print(f"missing-link: {rel_file} -> {target}")
                failures = 1

    if failures:
        return 1

    print("docs validation passed")
    return 0


if __name__ == "__main__":
    sys.exit(main())
