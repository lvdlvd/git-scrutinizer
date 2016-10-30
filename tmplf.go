package main

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	git "github.com/libgit2/git2go"
)

// This is the table of functions we want to have available in the html templates.
// Be sure to add only functions with no side effects here.  To change state,
// add a handler for a POST method in main.
var tmplFuncs = template.FuncMap{
	"timestamp": func(t time.Time) string { return t.UTC().Format(time.RFC3339) },
	"date":      func(t time.Time) string { return t.UTC().Format("2006-01-02") },
	"list":      func(vv ...interface{}) []interface{} { return vv },
	"join":      func(sep string, vals []string) string { return strings.Join(vals, sep) },
	"contains": func(list []string, val string) bool {
		for _, v := range list {
			if v == val {
				return true
			}
		}
		return false
	},
	"shortid": func(s string) string {
		if len(s) > 7 {
			return fmt.Sprintf("%s...", s[:7])
		}
		return s
	},
	"split":             func(sep, s string) []string { return strings.Split(s, sep) },    // note: reversed args
	"trimprefix":        func(pfx, s string) string { return strings.TrimPrefix(s, pfx) }, // note: reversed args
	"titlecase":         strings.Title,
	"git":               func() *git.Repository { return repository },
	"gitbranchall":      func(name string) (*git.Branch, error) { return repository.LookupBranch(name, git.BranchAll) },
	"gitbranchlocal":    func(name string) (*git.Branch, error) { return repository.LookupBranch(name, git.BranchLocal) },
	"gitbranchremote":   func(name string) (*git.Branch, error) { return repository.LookupBranch(name, git.BranchRemote) },
	"gitlog":            gitLog,
	"gitrefs":           gitRefNames,
	"gitnotes":          gitNotes,
	"gitconfig":         gitConfig,
	"gitdiffs":          gitDiffs,
	"gitpatches":        gitPatches,
	"gitdeltastring":    gitDeltaString,
	"gitdiffflagstring": gitDiffFlagString,
	"gittree":           gitTree,
	"gitblob":           gitBlob,
	"lineno":            func(i int) int { return i + 1 }, // no math in templates
	"gitnotesforfile":   gitNotesForFile,
}
