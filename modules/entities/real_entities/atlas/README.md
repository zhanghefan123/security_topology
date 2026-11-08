# Atlas path validation reproduction

The Atlas implementation is grouped by responsibility:

- `entities`: graph, path, and segment data structures.
- `init`: graph construction and initialization from topology or path files.
- `generate`: path generation and hierarchical segment generation.

Dependency direction: `generate` -> `init` -> `entities`.
