package main

import (
	"bufio"
	"io/ioutil"
	"net/textproto"
	"sort"
	"strings"
)

type Message struct {
	Header textproto.MIMEHeader
	Body   string
}

var headerNewlineToSpace = strings.NewReplacer("\n", " ", "\r", " ")

func (msg *Message) WriteTo(w *bufio.Writer) error {
	pw := textproto.NewWriter(w)

	var keys []string
	for k, _ := range msg.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		for _, v := range msg.Header[k] {
			v = headerNewlineToSpace.Replace(v)
			v = textproto.TrimString(v)
			if err := pw.PrintfLine("%s: %s", k, v); err != nil {
				return err
			}
		}
	}
	if err := pw.PrintfLine(""); err != nil {
		return err
	}

	dw := pw.DotWriter()
	if _, err := dw.Write([]byte(msg.Body)); err != nil {
		return err
	}
	return dw.Close()
}

func ReadMessage(r *bufio.Reader) (msg *Message, err error) {
	pr := textproto.NewReader(r)
	hdr, err := pr.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(pr.DotReader())
	if err != nil {
		return nil, err
	}

	return &Message{Header: hdr, Body: string(b)}, nil
}

// func getMessages() ([]*Message, error) {
// 	notes, err := gitNotesList(*ref)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var msgs []*Message

// 	for _, v := range notes {
// 		b, err := gitCatFile(v)
// 		if err != nil {
// 			return nil, err
// 		}
// 		r := bufio.NewReader(bytes.NewBuffer(b))
// 		for {
// 			msg, err := ReadMessage(r)
// 			if err == io.EOF {
// 				break
// 			}
// 			if err != nil {
// 				return nil, fmt.Errorf("Reading notes object %s: %v", v, err)
// 			}
// 			msgs = append(msgs, msg)
// 		}
// 	}

// 	log.Printf("Read %d messages on %d commits.", len(msgs), len(notes))
// 	return msgs, nil
// }
