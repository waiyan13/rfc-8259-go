# Vendored JSONTestSuite

- Source: https://github.com/nst/JSONTestSuite
- Commit: `1ef36fa01286573e846ac449e8683f8833c5b26a` (2024-11-22)
- Content: `test_parsing/` only (95 `y_`, 188 `n_`, 35 `i_` files) and `LICENSE` (MIT).
- Local changes: none. Do not edit the files. `go:embed` reads them as bytes.

To update the corpus:

1. Clone the upstream repository to a temporary directory.
2. Replace `test_parsing/` and `LICENSE` with the upstream copies.
3. Write the new commit and the new counts in this file.
4. Change the counts in `TestCorpusCounts`.
