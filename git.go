package main

import "github.com/libgit2/git2go"

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
	var names []string
	for {
		n, err := it.Next()
		if ge, ok := err.(*git.GitError); ok && ge.Code == git.ErrIterOver {
			break
		}
		if err != nil {
			return names, err
		}
		names = append(names, n)
	}
	return names, nil
}
