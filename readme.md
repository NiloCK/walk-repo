# WalkRepo

This packge exposes a single helper function, `WalkRepo`, which recreates `filepath.WalkDir` while also respecting encountered `.gitignore` configurations.

Useful in cases where you want some automated tooling to a git repository, especially where those repositories' directories are dominated by generated build artefacts (eg, `node_modules`).
