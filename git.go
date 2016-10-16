package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

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

func gitNotes() (map[string][]*Message, error) {
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
	ss := map[string][]*Message{}
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

		r := bufio.NewReader(bytes.NewBuffer(b.Contents()))
		for {
			msg, err := ReadMessage(r)
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("Reading notes object %s: %v", noteid, err)
			}
			ss[annid.String()] = append(ss[annid.String()], msg)
		}

	}
	return ss, nil
}

func gitNoteAppend(id *git.Oid, msg *Message) error {
	head, err := repository.Head()
	if err != nil {
		return err
	}
	branch, err := head.Branch().Name()
	if err != nil {
		return err
	}
	ref := path.Join(*refpfx, branch)

	sig, err := repository.DefaultSignature()
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	note, err := repository.Notes.Read(ref, id)
	if err == nil && note != nil {
		buf.WriteString(note.Message())
	}

	msg.Header.Set("Author", fmt.Sprintf("%s <%s>", sig.Name, sig.Email))
	msg.Header.Set("Date", sig.When.Format(time.RFC3339))
	msg.WriteTo(w)
	w.Flush()

	_, err = repository.Notes.Create(ref, sig, sig, id, buf.String(), true)
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

func gitPatches(ocid, ncid *git.Oid) ([]string, error) {
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
	var r []string
	for i := 0; i < N; i++ {
		dd, err := diff.Patch(i)
		if err != nil {
			return nil, err
		}
		s, err := dd.String()
		if err != nil {
			return nil, err
		}
		r = append(r, s)
	}
	return r, nil
}

func gitDeltaString(d git.Delta) string {
	switch d {
	case git.DeltaUnmodified:
		return "Unmodified"
	case git.DeltaAdded:
		return "Added"
	case git.DeltaDeleted:
		return "Deleted"
	case git.DeltaModified:
		return "Modified"
	case git.DeltaRenamed:
		return "Renamed"
	case git.DeltaCopied:
		return "Copied"
	case git.DeltaIgnored:
		return "Ignored"
	case git.DeltaUntracked:
		return "Untracked"
	case git.DeltaTypeChange:
		return "TypeChange"
	}
	return fmt.Sprintf("Delta[%d]", int(d))
}

func gitDiffFlagString(d git.DiffFlag) string {
	var f []string
	if d&git.DiffFlagBinary != 0 {
		f = append(f, "binary")
	}
	if d&git.DiffFlagNotBinary != 0 {
		f = append(f, "text")
	}
	if d&git.DiffFlagValidOid != 0 {
		f = append(f, "valid")
	}
	return strings.Join(f, ",")
}

func gitTree(path string) ([]*git.TreeEntry, error) {
	head, err := repository.Head()
	if err != nil {
		return nil, err
	}
	c, err := repository.LookupCommit(head.Target())
	if err != nil {
		return nil, err
	}
	tree, err := c.Tree()
	if err != nil {
		return nil, err
	}
	if path != "" {
		entry, err := tree.EntryByPath(path)
		if err != nil {
			return nil, err
		}
		if entry.Type != git.ObjectTree {
			return nil, fmt.Errorf("not a tree: %q", path)
		}
		tree, err = repository.LookupTree(entry.Id)
		if err != nil {
			return nil, err
		}
	}

	var r []*git.TreeEntry
	for i, e := uint64(0), tree.EntryCount(); i < e; i++ {
		r = append(r, tree.EntryByIndex(i))
	}
	return r, nil
}

func gitBlob(oid string) (<-chan string, error) {
	id, err := git.NewOid(oid)
	if err != nil {
		return nil, err
	}
	blob, err := repository.LookupBlob(id)
	if err != nil {
		return nil, err
	}
	r := bufio.NewScanner(bytes.NewReader(blob.Contents()))
	ch := make(chan string)
	go func() {
		defer close(ch)
		for r.Scan() {
			ch <- r.Text()
		}
		if err := r.Err(); err != nil {
			ch <- fmt.Sprintf("Error scanning %s: %v", oid, err)
		}
	}()
	return ch, nil
}
