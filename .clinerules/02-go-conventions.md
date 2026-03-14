# Go Code Conventions

## Style
- Use `log/slog` for all logging (structured logging). Never use `fmt.Println` for user output — use a thin `internal/ui` helper with `[INFO]`, `[WARN]`, `[ERROR]` prefixes.
- All exported types and functions MUST have doc comments.
- Use pointer receivers for structs with mutation methods.
- Use value receivers for small immutable structs.
- Name test files `*_test.go` in the same package (white-box testing).
- For interface-based testing, put interfaces in the consuming package, not the providing package.

## Error Handling
- Wrap errors with context: `fmt.Errorf("loading config %q: %w", name, err)`
- Use `errors.Is()` and `errors.As()` for checking, never `==`.
- Define sentinel errors as package-level `var` using `errors.New()`.

## File Creation Rules
- When creating a new `.go` file, ALWAYS start with the package declaration and imports.
- ALWAYS run `go build ./...` after creating or modifying a file to check for compilation errors.
- ALWAYS run `go vet ./...` after making changes.
- Run tests with `go test ./... -v` after completing each logical unit.

## Testing
- Table-driven tests where there are 3+ cases.
- Use `t.TempDir()` for any test that touches the filesystem.
- Use `t.Helper()` in test helper functions.
- Name test functions: `TestTypeName_MethodName_condition`.