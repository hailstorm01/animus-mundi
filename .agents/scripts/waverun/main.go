// Command waverun builds one wave unattended (lead-architect Runbook D7).
//
// It opens one fresh headless lead-architect session per task. Each session
// moves one task and merges it into an autoloop/ run branch, never the
// integration branch. After each session the lead adds one row to the run
// file, and its last cell decides what happens next: CONTINUE, WAVE-COMPLETE
// or STOP: <reason>. The runner also stops on its own budgets (tasks, spend
// per session, hours), on a failed session, on any refused tool call, and on
// a missing or extra run row. It never merges the run branch: the user does.
//
//	go tool task wave-run -- -Wave B -Tasks 3
//	go tool task wave-run -- -Wave B -DryRun
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	wave          = flag.String("Wave", "", "wave letter (required)")
	tasks         = flag.Int("Tasks", 3, "most sessions to run")
	maxUsdPerTask = flag.Float64("MaxUsdPerTask", 40, "spend budget per session, passed to --max-budget-usd")
	maxHours      = flag.Float64("MaxHours", 8, "no new session starts after this many hours")
	milestone     = flag.String("Milestone", "m2", "milestone folder under .agents/tasks")
	base          = flag.String("Base", "main", "integration branch the run branch starts from")
	effort        = flag.String("Effort", "", "--effort for each session, if set")
	claudePath    = flag.String("Claude", "", "path to the claude CLI, if it is not on PATH")
	dryRun        = flag.Bool("DryRun", false, "print what would run, then exit")
)

// run is the state Stop reports on.
type run struct {
	branch, runFile, logDir string
	started                 time.Time
	spent                   float64
}

// result is the part of `claude -p --output-format json` the runner reads.
type result struct {
	IsError           bool    `json:"is_error"`
	Subtype           string  `json:"subtype"`
	TotalCostUSD      float64 `json:"total_cost_usd"`
	SessionID         string  `json:"session_id"`
	NumTurns          int     `json:"num_turns"`
	TerminalReason    string  `json:"terminal_reason"`
	APIErrorStatus    int     `json:"api_error_status"`
	APIError          string  `json:"api_error"`
	Result            string  `json:"result"`
	PermissionDenials []struct {
		ToolName  string          `json:"tool_name"`
		ToolInput json.RawMessage `json:"tool_input"`
	} `json:"permission_denials"`
}

func main() {
	flag.Parse()
	if *wave == "" {
		fmt.Println("-Wave is required, e.g. -Wave B")
		os.Exit(2)
	}
	top, err := git("rev-parse", "--show-toplevel")
	if err != nil {
		fmt.Println("not inside a git checkout")
		os.Exit(1)
	}
	if err := os.Chdir(strings.TrimSpace(top)); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	branch := fmt.Sprintf("autoloop/%s-wave-%s", *milestone, strings.ToLower(*wave))
	slug := strings.ReplaceAll(branch, "/", "-")
	r := &run{
		branch:  branch,
		runFile: fmt.Sprintf(".agents/tasks/%s/runs/%s.md", *milestone, slug),
		logDir:  filepath.Join(os.TempDir(), "wave-run", slug),
		started: time.Now(),
	}

	exe := findClaude()
	if exe == "" {
		fmt.Println("claude CLI not found; pass -Claude <path to claude.exe>")
		os.Exit(1)
	}
	if !branchExists(*base) {
		fmt.Printf("Base branch %s does not exist\n", *base)
		os.Exit(1)
	}

	cliArgs := []string{
		"--model", "opus",
		"--permission-mode", "auto",
		"--permission-prompts", "none",
		"--max-budget-usd", strconv.FormatFloat(*maxUsdPerTask, 'f', -1, 64),
		"--output-format", "json",
	}
	if *effort != "" {
		cliArgs = append(cliArgs, "--effort", *effort)
	}
	cliArgs = append(cliArgs, "--disallowedTools", "Bash(git push *),PowerShell(git push *)")

	if *dryRun {
		fmt.Printf("Run branch: %s (exists: %t)\n", r.branch, branchExists(r.branch))
		fmt.Printf("Run file:   %s (%d rows)\n", r.runFile, len(runRows(r.runText())))
		fmt.Printf("Claude:     %s\n", exe)
		fmt.Printf("Session:    claude -p \"@lead-architect next (wave run: ...)\" %s\n", strings.Join(cliArgs, " "))
		os.Exit(0)
	}
	keepAwake()

	if !branchExists(r.branch) {
		if _, err := git("branch", r.branch, *base); err != nil {
			fmt.Printf("Could not create %s\n", r.branch)
			os.Exit(1)
		}
		fmt.Printf("Created %s from %s.\n", r.branch, *base)
	}
	if err := os.MkdirAll(r.logDir, 0o755); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for n := 1; n <= *tasks; n++ {
		if time.Since(r.started).Hours() >= *maxHours {
			r.stop(fmt.Sprintf("time budget of %g h used up", *maxHours), 1)
		}

		before := len(runRows(r.runText()))
		session := fmt.Sprintf("%s-%d", time.Now().Format("20060102-150405"), n)
		// A headless session has no scratch folder of its own. Left to choose,
		// the lead put its worktrees under /tmp, which is C:\tmp here.
		scratch := filepath.Join(r.logDir, session)
		if err := os.MkdirAll(scratch, 0o755); err != nil {
			r.stop(fmt.Sprintf("could not create the scratch folder: %v", err), 1)
		}
		prompt := fmt.Sprintf("@lead-architect next (wave run: wave %s, run branch %s, run file %s, scratch %s)", *wave, r.branch, r.runFile, filepath.ToSlash(scratch))
		logPath := filepath.Join(r.logDir, session+".json")
		fmt.Printf("[%d/%d] %s lead session on %s (log: %s)\n", n, *tasks, time.Now().Format("15:04"), r.branch, logPath)

		// The result JSON's duration_ms covers only a session's last segment
		// when it waited on a background command, so time it here.
		began := time.Now()
		cmd := exec.Command(exe, append([]string{"-p", prompt, "-n", fmt.Sprintf("wave-run %s #%d", *wave, n)}, cliArgs...)...)
		cmd.Stderr = os.Stderr
		out, runErr := cmd.Output()
		took := time.Since(began).Round(time.Second)
		if err := os.WriteFile(logPath, out, 0o644); err != nil {
			fmt.Printf("   could not write the session log: %v\n", err)
		}
		// Remove fails on a folder that isn't empty, which is left for the user.
		if os.Remove(scratch) != nil {
			fmt.Printf("   scratch folder not empty, left in place: %s\n", scratch)
		}

		var res result
		parsed := json.Unmarshal(out, &res) == nil
		if parsed {
			r.spent += res.TotalCostUSD
		}
		fmt.Printf("   took %s, $%.2f\n", took, res.TotalCostUSD)
		if runErr != nil || !parsed || res.IsError || len(res.PermissionDenials) > 0 {
			for _, l := range diagnosis(res, parsed, out, runErr) {
				fmt.Println("   " + l)
			}
		}
		if runErr != nil || !parsed {
			r.stop(fmt.Sprintf("session %d exited with code %d", n, exitCode(runErr)), 1)
		}
		if res.IsError {
			r.stop(fmt.Sprintf("session %d ended with %s", n, res.Subtype), 1)
		}
		if len(res.PermissionDenials) > 0 {
			r.stop(fmt.Sprintf("session %d had refused tool calls", n), 1)
		}

		text := r.runText()
		rows := runRows(text)
		if len(rows) != before+1 {
			r.stop(fmt.Sprintf("session %d added %d run rows, expected 1", n, len(rows)-before), 1)
		}
		row := rows[len(rows)-1]
		fmt.Printf("   %s\n", row)

		switch sig := signal(row); {
		case sig == "CONTINUE":
			continue
		case sig == "WAVE-COMPLETE":
			pending, err := r.pendingCards(text)
			if err != nil {
				r.stop(fmt.Sprintf("lead says complete, but the card check failed: %v", err), 1)
			}
			if len(pending) > 0 {
				r.stop("lead says complete, but still not executable: "+strings.Join(pending, "; "), 1)
			}
			fmt.Println("   Verifying gates on the run branch...")
			if !r.gatesPass() {
				r.stop("lead says complete, but check or sim fails on the run branch", 1)
			}
			r.stop(fmt.Sprintf("wave %s complete", *wave), 0)
		case strings.HasPrefix(sig, "STOP"):
			r.stop("lead: "+sig, 1)
		default:
			r.stop(fmt.Sprintf("unknown signal '%s'", sig), 1)
		}
	}
	r.stop(fmt.Sprintf("task budget of %d used up", *tasks), 0)
}

func (r *run) stop(reason string, code int) {
	hours := time.Since(r.started).Hours()
	fmt.Println()
	fmt.Printf("== Wave run stopped: %s\n", reason)
	fmt.Printf("   Spent: $%.2f over %.2f h. Session logs: %s\n", r.spent, hours, r.logDir)
	fmt.Printf("   Review:  git log --oneline %s..%s\n", *base, r.branch)
	fmt.Printf("            git show %s:%s\n", r.branch, r.runFile)
	fmt.Printf("   Land it: git merge --no-ff %s   (on %s, when you are satisfied)\n", r.branch, *base)
	fmt.Printf("   Resume:  go tool task wave-run -- -Wave %s -Tasks <n>\n", *wave)
	os.Exit(code)
}

// diagnosis says why a session failed: the exit error, the result's error
// fields and refused tool calls, or the end of the output when it is not a
// result. A usage limit, for example, exits 1 with is_error set, a 429 in
// api_error_status and the reset time in result.
func diagnosis(res result, parsed bool, out []byte, runErr error) []string {
	var ls []string
	if runErr != nil {
		ls = append(ls, fmt.Sprintf("exit: %v", runErr))
	}
	if !parsed {
		tail := strings.TrimSpace(string(out))
		if tail == "" {
			return append(ls, "no output on stdout (stderr, if any, is above)")
		}
		if len(tail) > 1500 {
			tail = "..." + tail[len(tail)-1500:]
		}
		return append(ls, "output is not a result JSON; it ends:", tail)
	}
	ls = append(ls, fmt.Sprintf("session %s: subtype %s, terminal reason %s, is_error %t, %d turns",
		res.SessionID, res.Subtype, res.TerminalReason, res.IsError, res.NumTurns))
	if res.APIErrorStatus != 0 || res.APIError != "" {
		ls = append(ls, fmt.Sprintf("API error: %d %s", res.APIErrorStatus, res.APIError))
	}
	if msg := strings.TrimSpace(res.Result); msg != "" {
		if len(msg) > 500 {
			msg = msg[:500] + "..."
		}
		ls = append(ls, "result: "+msg)
	}
	for _, d := range res.PermissionDenials {
		ls = append(ls, fmt.Sprintf("refused: %s %s", d.ToolName, d.ToolInput))
	}
	return ls
}

// runText is the run file as committed on the run branch, or nothing before
// the first row.
func (r *run) runText() []string {
	if !branchExists(r.branch) {
		return nil
	}
	if ls, err := git("ls-tree", "--name-only", r.branch, "--", r.runFile); err != nil || strings.TrimSpace(ls) == "" {
		return nil
	}
	text, err := git("show", r.branch+":"+r.runFile)
	if err != nil {
		return nil
	}
	return lines(text)
}

// pendingCards returns the run file header's cards still listed in
// NotYetExecutable on the run branch.
func (r *run) pendingCards(text []string) ([]string, error) {
	names, ok := cardsLine(text)
	if !ok {
		return []string{"(run file has no Cards line)"}, nil
	}
	src, err := git("show", r.branch+":pkg/engine/cards/cards.go")
	if err != nil {
		return nil, err
	}
	listed, err := notYetExecutable(src)
	if err != nil {
		return nil, err
	}
	return pending(names, listed), nil
}

// gatesPass runs check and sim on the run branch in a temporary worktree.
func (r *run) gatesPass() bool {
	wt := filepath.Join(os.TempDir(), fmt.Sprintf("wave-run-verify-%d", os.Getpid()))
	if _, err := git("worktree", "add", "--detach", wt, r.branch); err != nil {
		return false
	}
	defer git("worktree", "remove", "--force", wt)
	for _, gate := range []string{"check", "sim"} {
		cmd := exec.Command("go", "-C", wt, "tool", "task", gate)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if cmd.Run() != nil {
			return false
		}
	}
	return true
}

var rowRe = regexp.MustCompile(`^\|\s*\d+\s*\|`)

// runRows returns the run table's numbered rows.
func runRows(text []string) []string {
	var rows []string
	for _, l := range text {
		if rowRe.MatchString(l) {
			rows = append(rows, l)
		}
	}
	return rows
}

// signal returns a run row's last cell. An escaped pipe (\|) is part of the
// cell's text, as when a STOP reason quotes a piped command.
func signal(row string) string {
	row = strings.TrimSuffix(strings.TrimSpace(row), "|")
	i := len(row) - 1
	for ; i >= 0; i-- {
		if row[i] == '|' && (i == 0 || row[i-1] != '\\') {
			break
		}
	}
	return strings.TrimSpace(strings.ReplaceAll(row[i+1:], `\|`, "|"))
}

// cardsLine returns the names on the header's Cards line.
func cardsLine(text []string) ([]string, bool) {
	const prefix = "- **Cards:**"
	for _, l := range text {
		if !strings.HasPrefix(l, prefix) {
			continue
		}
		var names []string
		for _, n := range strings.Split(strings.TrimPrefix(l, prefix), ";") {
			if n = strings.TrimSpace(n); n != "" {
				names = append(names, n)
			}
		}
		return names, true
	}
	return nil, false
}

// notYetExecutable returns the string literals of the NotYetExecutable slice
// in cards.go source.
func notYetExecutable(src string) ([]string, error) {
	f, err := parser.ParseFile(token.NewFileSet(), "cards.go", src, 0)
	if err != nil {
		return nil, err
	}
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			continue
		}
		for _, spec := range gd.Specs {
			vs := spec.(*ast.ValueSpec)
			for i, name := range vs.Names {
				if name.Name != "NotYetExecutable" || i >= len(vs.Values) {
					continue
				}
				lit, ok := vs.Values[i].(*ast.CompositeLit)
				if !ok {
					return nil, fmt.Errorf("NotYetExecutable is not a slice literal")
				}
				var names []string
				for _, elt := range lit.Elts {
					bl, ok := elt.(*ast.BasicLit)
					if !ok || bl.Kind != token.STRING {
						return nil, fmt.Errorf("NotYetExecutable holds a non-string element")
					}
					s, err := strconv.Unquote(bl.Value)
					if err != nil {
						return nil, err
					}
					names = append(names, s)
				}
				return names, nil
			}
		}
	}
	return nil, fmt.Errorf("cards.go has no NotYetExecutable")
}

// pending returns the names that are also listed, in the order given.
func pending(names, listed []string) []string {
	in := make(map[string]bool, len(listed))
	for _, l := range listed {
		in[l] = true
	}
	var out []string
	for _, n := range names {
		if in[n] {
			out = append(out, n)
		}
	}
	return out
}

func findClaude() string {
	if *claudePath != "" {
		return *claudePath
	}
	if p, err := exec.LookPath("claude"); err == nil {
		return p
	}
	// The IDE extensions ship the CLI without putting it on PATH.
	home, _ := os.UserHomeDir()
	bin := "claude"
	if runtime.GOOS == "windows" {
		bin = "claude.exe"
	}
	var found []string
	for _, ext := range []string{".antigravity-ide", ".vscode"} {
		m, _ := filepath.Glob(filepath.Join(home, ext, "extensions", "anthropic.claude-code-*", "resources", "native-binary", bin))
		found = append(found, m...)
	}
	sort.Slice(found, func(i, j int) bool { return modTime(found[i]).After(modTime(found[j])) })
	if len(found) > 0 {
		return found[0]
	}
	return ""
}

func modTime(path string) time.Time {
	// The extension folder's time, as the PowerShell runner sorted by.
	fi, err := os.Stat(filepath.Dir(filepath.Dir(filepath.Dir(path))))
	if err != nil {
		return time.Time{}
	}
	return fi.ModTime()
}

func branchExists(name string) bool {
	_, err := git("rev-parse", "--verify", "--quiet", "refs/heads/"+name)
	return err == nil
}

func git(args ...string) (string, error) {
	out, err := exec.Command("git", args...).Output()
	return string(out), err
}

func lines(s string) []string {
	ls := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	if len(ls) > 0 && ls[len(ls)-1] == "" {
		ls = ls[:len(ls)-1]
	}
	return ls
}

func exitCode(err error) int {
	if ee, ok := err.(*exec.ExitError); ok {
		return ee.ExitCode()
	}
	if err != nil {
		return -1
	}
	return 0
}
