# x/chardet

`x/chardet` is a pure-Go character encoding detector inspired by `chardetng`.

## Upstream base

This rewrite is based on `chardetng` commit:

- `3b893dbbd06098ce84144575fbd70dea73a576eb`
- with patch `47e68d9c65ea4f3063ee3dd8f5f4f912f3bb7076` (`xhorak/chardetng`: Add support to GB18030 encoding detection)

## API

- `NewDetector() *Detector`
- `(*Detector).Feed(chunk []byte, last bool) bool`
- `(*Detector).Guess(tld []byte, allowUTF8 bool) Encoding`
- `(*Detector).GuessAssess(tld []byte, allowUTF8 bool) GuessResult`
- `TLDMayAffectGuess(tld []byte) bool`

`tld` is expected to be the right-most DNS label in lower-case ASCII (e.g. `"jp"`, `"com"`, or punycode like `"xn--..."`).

## Data conversion

CJK frequent-character tables are generated from `chardetng/src/data.rs`:

```bash
go run ./x/chardet/tools/gen_cjk_data
```

This generates:

- `x/chardet/cjk_data_gen.go`
- `x/chardet/singlebyte_data_gen.go` via `go run ./x/chardet/tools/gen_singlebyte_data`
- `x/chardet/tld_data_gen.go` via `go run ./x/chardet/tools/gen_tld_data`

## Benchmark

Use package benchmarks for local hot-path performance:

```bash
GOCACHE=/tmp/go-cache go test ./x/chardet -run ^$ -bench BenchmarkDetectorGuessMixed -benchmem
GOCACHE=/tmp/go-cache go test ./x/chardet -run ^$ -bench BenchmarkScoreDecodedByEncoding -benchmem
```

Current benchmark focus:

- `BenchmarkDetectorGuessMixed`: end-to-end `Reset + Feed + Guess` path.
- `BenchmarkScoreDecodedByEncoding`: per-encoding scoring cost breakdown.

## Compare (Go vs Rust vs saintfish)

Build Rust detector helper first:

```bash
cd x/chardet/tools/compare/tools/rust_detect_tsv
cargo build --release
```

Notes:

- This Rust helper is for local compare workflow only.
- It does not affect the publishable Go package dependencies.

Generate dataset:

```bash
cd x/chardet/tools/compare
go run . generate -out dataset.tsv -repeat 30
```

Run cross-implementation compare:

```bash
cd x/chardet/tools/compare
GOCACHE=/tmp/go-cache go run . compare
```

Outputs:

- `x/chardet/tools/compare/last_report.txt` (default `last_report.txt` in compare dir)
- `x/chardet/tools/compare/history_report.log` (default `history_report.log` in compare dir)

Metrics reported:

- `exact`: exact encoding-label match ratio
- `family`: encoding-family match ratio
- `throughput`: samples per second

## Optimization Methodology

Use this order for performance work:

1. `go test` must stay green first.
2. Run benchmarks to locate hot paths.
3. Optimize one hotspot at a time.
4. Re-run compare and verify:
   - no regression in `exact` / `family`
   - throughput improves or stays acceptable
5. Keep changes if and only if quality target is preserved.

Recommended acceptance gate:

- `exact` and `family` should not regress against current baseline (`exact=0.9091`, `family=1.0000` on the default compare dataset).

## Regression Correction Note

An aggressive optimization was tried once: CJK byte-shape fast-path early return in `scoreAll`.
It improved speed but caused non-CJK false positives (example: `ibm866 -> shift-jis`) and reduced compare accuracy.

Decision:

- remove this fast-path and keep the safer path.
- prefer conservative optimizations (allocation reduction, decoder-path micro-optimizations) over heuristic pruning that changes candidate coverage.
