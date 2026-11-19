// The keyword set: the standalone keywords a card prints or is granted (FR-KWD-002), their printed names, and how two sets combine.

package ability

import (
	"regexp"
	"strconv"
	"strings"
)

// Keywords holds the standalone keyword values printed on (or granted to) an
// entity. Numeric keywords use 0 for "absent"; Resistant may be negative.
type Keywords struct {
	Resistant        int  `json:"resistant,omitempty"`
	Stalwart         bool `json:"stalwart,omitempty"`
	Aggressive       int  `json:"aggressive,omitempty"`
	Ethereal         bool `json:"ethereal,omitempty"`
	Ward             bool `json:"ward,omitempty"`
	Versatile        bool `json:"versatile,omitempty"`
	Guardian         bool `json:"guardian,omitempty"`
	Armor            int  `json:"armor,omitempty"`
	DeploysExhausted bool `json:"deploysExhausted,omitempty"`
}

// keywordDef is one standalone keyword: its printed name and its Keywords
// field. A numeric keyword (num) prints its N, as in "Resistant 2"; a flag
// keyword prints its name alone.
type keywordDef struct {
	name string
	num  func(*Keywords) *int
	flag func(*Keywords) *bool
}

// keywordDefs lists the standalone keywords in printed order. ParseKeywordLine,
// Keywords.Lines and Keywords.Add all walk it, so each keyword's name is
// written only here, or, for a keyword that is a triggered ability, in
// KeywordTrigger.String.
var keywordDefs = []keywordDef{
	{name: "Resistant", num: func(k *Keywords) *int { return &k.Resistant }},
	{name: "Stalwart", flag: func(k *Keywords) *bool { return &k.Stalwart }},
	{name: KeywordTriggerAggressive.String(), num: func(k *Keywords) *int { return &k.Aggressive }},
	{name: "Ethereal", flag: func(k *Keywords) *bool { return &k.Ethereal }},
	{name: "Ward", flag: func(k *Keywords) *bool { return &k.Ward }},
	{name: "Versatile", flag: func(k *Keywords) *bool { return &k.Versatile }},
	{name: "Guardian", flag: func(k *Keywords) *bool { return &k.Guardian }},
	{name: "Armor", num: func(k *Keywords) *int { return &k.Armor }},
	{name: "Deploys Exhausted", flag: func(k *Keywords) *bool { return &k.DeploysExhausted }},
}

// Lines lists k's standalone keywords in printed form and order, never nil.
func (k Keywords) Lines() []string {
	out := []string{}
	for _, d := range keywordDefs {
		switch {
		case d.num != nil:
			if n := *d.num(&k); n != 0 {
				out = append(out, d.name+" "+strconv.Itoa(n))
			}
		case *d.flag(&k):
			out = append(out, d.name)
		}
	}
	return out
}

// Add folds o into k: a flag is set, a number is added. It walks keywordDefs,
// so a later keyword needs no new line here.
func (k *Keywords) Add(o Keywords) {
	for _, d := range keywordDefs {
		switch {
		case d.num != nil:
			*d.num(k) += *d.num(&o)
		case d.flag != nil:
			*d.flag(k) = *d.flag(k) || *d.flag(&o)
		}
	}
}

// keywordN matches a numeric keyword's N, which always prints.
var keywordN = regexp.MustCompile(`^-?\d+$`)

// ParseKeywordLine parses one standalone keyword line from rules text
// (keywordDefs).
func ParseKeywordLine(line string, kw *Keywords) bool {
	for _, d := range keywordDefs {
		if d.flag != nil {
			if line == d.name {
				*d.flag(kw) = true
				return true
			}
			continue
		}
		if n, ok := strings.CutPrefix(line, d.name+" "); ok && keywordN.MatchString(n) {
			v, _ := strconv.Atoi(n)
			*d.num(kw) += v
			return true
		}
	}
	return false
}
