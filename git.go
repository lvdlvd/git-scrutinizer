package main

import (
	"fmt"
	"path"

	git "github.com/libgit2/git2go"
)

func gitConfig() ([]*git.ConfigEntry, error) {
	cfg, err := repository.Config()
	if err != nil {
		return nil, err
	}

	it, err := cfg.NewIterator()
	if err != nil {
		return nil, err
	}
	var ss []*git.ConfigEntry
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

func gitNoteAppend(id *git.Oid, msg string) error {
	head, err := repository.Head()
	if err != nil {
		return err
	}
	branch, err := head.Branch().Name()
	if err != nil {
		return err
	}
	ref := path.Join(*refpfx, branch)

	note, err := repository.Notes.Read(ref, id)
	if err == nil && note != nil {
		msg = fmt.Sprintf("%s\n%s", note.Message(), msg)
	}

	sig, err := repository.DefaultSignature()
	if err != nil {
		return err
	}

	_, err = repository.Notes.Create(ref, sig, sig, id, msg, true)
	return err
}

func gitDiffs(ocid, ncid *git.Oid) ([]*git.DiffDelta, error) {
	oc, err := repository.LookupCommit(ocid)
	if err != nil {
		return nil, err
	}
	otree, err := oc.Tree()
	if err != nil {
		return nil, err
	}

	nc, err := repository.LookupCommit(ncid)
	if err != nil {
		return nil, err
	}
	ntree, err := nc.Tree()
	if err != nil {
		return nil, err
	}
	opts, err := git.DefaultDiffOptions()
	if err != nil {
		return nil, err
	}
	diff, err := repository.DiffTreeToTree(otree, ntree, &opts)
	if err != nil {
		return nil, err
	}
	N, err := diff.NumDeltas()
	if err != nil {
		return nil, err
	}
	var r []*git.DiffDelta
	for i := 0; i < N; i++ {
		dd, err := diff.GetDelta(i)
		if err != nil {
			return nil, err
		}
		r = append(r, &dd)
	}
	return r, nil
}
