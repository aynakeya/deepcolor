# chardet compare

This toolchain compares:

- migrated `x/chardet`
- upstream Rust `chardetng`
- `github.com/saintfish/chardet` (for reference only)

on the same generated dataset.

## 1) Build Rust detector binary

`detect_tsv` is an independent local dev tool and is not part of the publishable Go package.

Build from repository root:

```bash
cd x/chardet/tools/compare/tools/rust_detect_tsv
cargo build --release
```

## 2) Generate dataset

```bash
cd x/chardet/tools/compare
go run . generate -out dataset.tsv -repeat 30
```

## 3) Run comparison (accuracy + speed + report files)

```bash
cd x/chardet/tools/compare
go run . compare \
  -data dataset.tsv \
  -rust-bin tools/rust_detect_tsv/target/release/detect_tsv \
  -rust-out rust_out.tsv \
  -report last_report.txt \
  -history history_report.log
```

Output includes:

- exact accuracy
- family accuracy
- runtime / throughput
- top mismatch pairs for all three engines

Files:

- `-report`: overwritten each run (latest snapshot)
- `-history`: append-only log (with timestamp per run)
