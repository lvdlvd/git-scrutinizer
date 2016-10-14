package main

import (
	"html/template"
	"strings"
	"time"

	git "github.com/libgit2/git2go"
)

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
	"split":      func(sep, s string) []string { return strings.Split(s, sep) },    // note: reversed args
	"trimprefix": func(pfx, s string) string { return strings.TrimPrefix(s, pfx) }, // note: reversed args
	"titlecase":  strings.Title,
	"git":        func() *git.Repository { return repository },
	"gitlog":     gitLog,
}
