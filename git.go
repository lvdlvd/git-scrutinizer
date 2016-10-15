package main

import (
	"path"

	git "github.com/libgit2/git2go"
)

// log of master..HEAD
func gitLog() ([]*git.Commit, error) {
	w, err := repository.Walk()
	if err != nil {
		return nil, err
	}

	w.Sorting(git.SortTopological | git.SortTime)
	if err := w.PushHead(); err != nil {
		return nil, err
	}
	if err := w.HideRef("refs/heads/master"); err != nil {
		return nil, err
	}

	var commits []*git.Commit

	if err := w.Iterate(func(commit *git.Commit) bool {
		commits = append(commits, commit)
		return true
	}); err != nil {
		return nil, err
	}

	return commits, nil
}

func gitRefNames() ([]string, error) {
	it, err := repository.NewReferenceNameIterator()
	if err != nil {
		return nil, err
	}
	var ss []string
	for {
		n, err := it.Next()
		if ge, ok := err.(*git.GitError); ok && ge.Code == git.ErrIterOver {
			break
		}
		if err != nil {
			return ss, err
		}
		ss = append(ss, n)
	}
	return ss, nil
}

func gitNotes() (map[string]string, error) {
	head, err := repository.Head()
	if err != nil {
		return nil, err
	}
	branch, err := head.Branch().Name()
	if err != nil {
		return nil, err
	}
	it, err := repository.NewNoteIterator(path.Join(*refpfx, branch))
	if ge, ok := err.(*git.GitError); ok && ge.Code == git.ErrNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ss := map[string]string{}
	for {
		noteid, annid, err := it.Next()
		if ge, ok := err.(*git.GitError); ok && ge.Code == git.ErrIterOver {
			break
		}
		if err != nil {
			return ss, err
		}
		obj, err := repository.Lookup(noteid)
		if err != nil {
			return ss, err
		}
		b, err := obj.AsBlob()
		if err != nil {
			return ss, err
		}
		ss[annid.String()] = string(b.Contents())
	}
	return ss, nil
}
