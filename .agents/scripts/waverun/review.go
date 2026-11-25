package main

import (
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"
)

// A complete wave gets one structural review (lead-architect Runbook E): a
// headless session that commits a report to the run branch and nothing else.
// The user decides its findings in an interactive session, and no other wave
// starts until they are decided.

const (
	reportSuffix = "-structural-review.md"
	reviewID     = "review"
)

// report is one structural review report as committed on branch.
type report struct {
	path, branch string
	status       string // "open" or "decided (<date>)"; empty if the line is missing
	findings     string // "5 (2 duplicates)"
}

var (
	statusRe   = regexp.MustCompile(`(?m)^- \*\*Status:\*\*[ \t]*(.*?)[ \t\r]*$`)
	findingsRe = regexp.MustCompile(`(?m)^- \*\*Findings:\*\*[ \t]*(.*?)[ \t\r]*$`)
	noneRe     = regexp.MustCompile(`^0\b`)
	countRe    = regexp.MustCompile(`^\d+\b`)
)

func parseReport(branch, path, text string) report {
	rep := report{path: path, branch: branch}
	if m := statusRe.FindStringSubmatch(text); m != nil {
		rep.status = m[1]
	}
	if m := findingsRe.FindStringSubmatch(text); m != nil {
		rep.findings = m[1]
	}
	return rep
}

// decided is true only when the status says so: a report without a readable
// status still waits for the user.
func (rep report) decided() bool { return strings.HasPrefix(rep.status, "decided") }

// problem says what is wrong with a report the review session just wrote, or
// nothing. Only a review with no findings may mark itself decided.
func (rep report) problem() string {
	switch none := noneRe.MatchString(rep.findings); {
	case !countRe.MatchString(rep.findings):
		return "its Findings line has no count"
	case none && !rep.decided():
		return "it has no findings but isn't marked decided"
	case !none && rep.status != "open":
		return fmt.Sprintf("it has findings but its status is %q, not \"open\"", rep.status)
	}
	return ""
}

func (rep report) read() string { return fmt.Sprintf("git show %s:%s", rep.branch, rep.path) }

func (rep report) decide() string {
	return fmt.Sprintf("@lead-architect decide structural review (branch %s, report %s)", rep.branch, rep.path)
}

// undecided returns the reports no branch has decided, one per path, as on
// the first branch that has it. own is the starting wave's report, which
// never blocks its own wave.
func undecided(reps []report, own string) []report {
	done := map[string]bool{own: true}
	for _, rep := range reps {
		if rep.decided() {
			done[rep.path] = true
		}
	}
	var out []report
	for _, rep := range reps {
		if !done[rep.path] {
			done[rep.path] = true
			out = append(out, rep)
		}
	}
	return out
}

// allReports reads every report on the base branch, then on each autoloop/
// branch.
func allReports(base string) ([]report, error) {
	out, err := git("for-each-ref", "--format=%(refname:short)", "refs/heads/autoloop/")
	if err != nil {
		return nil, fmt.Errorf("git for-each-ref: %w", err)
	}
	var reps []report
	for _, branch := range append([]string{base}, strings.Fields(out)...) {
		ls, err := git("ls-tree", "-r", "--name-only", branch, "--", ".agents/tasks/")
		if err != nil {
			return nil, fmt.Errorf("git ls-tree %s: %w", branch, err)
		}
		for _, p := range lines(ls) {
			if path.Base(path.Dir(p)) != "runs" || !strings.HasSuffix(p, reportSuffix) {
				continue
			}
			text, err := git("show", branch+":"+p)
			if err != nil {
				return nil, fmt.Errorf("git show %s:%s: %w", branch, p, err)
			}
			reps = append(reps, parseReport(branch, p, text))
		}
	}
	return reps, nil
}

// report reads the run's own report from the run branch.
func (r *run) report() (report, bool) {
	text, err := git("show", r.branch+":"+r.reportFile)
	if err != nil {
		return report{}, false
	}
	return parseReport(r.branch, r.reportFile, text), true
}

// review runs the structural review of a complete wave and sets what stop
// prints about it. The wave is complete whatever happens here.
func (r *run) review(exe string, cliArgs []string) {
	if rep, ok := r.report(); ok {
		// A resumed run: the wave was reviewed when it first completed.
		r.note(rep)
		return
	}
	if time.Since(r.started).Hours() >= *maxHours {
		r.reviewFailed("the time budget is used up")
		return
	}
	before, err := git("rev-parse", r.branch)
	if err != nil {
		r.reviewFailed("could not read the run branch")
		return
	}
	fmt.Println("   Structural review of the wave...")
	s, err := r.session(exe, cliArgs, reviewID, func(scratch string) string {
		return fmt.Sprintf("@lead-architect structural review (wave run: wave %s, run branch %s, base %s, report %s, scratch %s). Your instructions are in .agents/skills/lead-architect/SKILL.md; read that first.", *wave, r.branch, *base, r.reportFile, scratch)
	})
	if err != nil {
		r.reviewFailed(err.Error())
		return
	}
	if rep, ok := r.report(); ok {
		r.note(rep)
		if p := rep.problem(); p != "" {
			r.undecided = true
			r.warn("check the report: " + p)
			if rep.decided() {
				r.notes = append(r.notes, "Decide:  "+rep.decide())
			}
		}
	} else {
		r.reviewFailed("the session committed no report")
	}
	if !s.clean() {
		r.warn("the review session did not end cleanly (see above)")
	}
	changed, err := git("diff", "--name-only", strings.TrimSpace(before), r.branch)
	if err != nil {
		r.warn("could not check which files the review changed")
	}
	if extra := except(lines(changed), r.reportFile); len(extra) > 0 {
		r.warn("the review changed files besides its report: " + strings.Join(extra, "; "))
	}
}

// note sets what stop prints about the wave's report.
func (r *run) note(rep report) {
	if rep.decided() {
		r.notes = []string{fmt.Sprintf("Structural review findings: %s, %s", rep.findings, rep.status)}
		return
	}
	r.undecided = true
	r.notes = []string{
		fmt.Sprintf("Structural review findings: %s, none decided", rep.findings),
		"Read:    " + rep.read(),
		"Decide:  " + rep.decide(),
	}
}

// reviewFailed leaves the review to the user.
func (r *run) reviewFailed(why string) {
	r.notes = []string{
		"No structural review: " + why,
		fmt.Sprintf("Run one: @lead-architect structural review (wave %s, branch %s)", *wave, r.branch),
	}
}

func (r *run) warn(s string) { r.notes = append(r.notes, "Warning: "+s) }

// blocked says which undecided reviews hold back a new wave.
func blocked(reps []report) []string {
	var ls []string
	for _, rep := range reps {
		ls = append(ls,
			fmt.Sprintf("Undecided structural review: %s on %s, findings: %s", rep.path, rep.branch, rep.findings),
			"   Read:    "+rep.read(),
			"   Decide:  "+rep.decide())
	}
	return ls
}

// except returns the names other than one.
func except(names []string, one string) []string {
	var out []string
	for _, n := range names {
		if n != one {
			out = append(out, n)
		}
	}
	return out
}
