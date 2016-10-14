package main

import (
	"bufio"
	"bytes"
	"net/textproto"
	"reflect"
	"testing"
)

func TestMessageWrite(t *testing.T) {

	msg := Message{Header: textproto.MIMEHeader{}}
	msg.Header.Add("Author", "Luuk van Dijk <lvd@daedalean.ai>")
	msg.Header.Add("Status", "resolved")
	msg.Body = `This is the body
of the message it has

multiple lines
and also
Aheader: that is not real

.
and a dot

`
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	if err := msg.WriteTo(w); err != nil {
		t.Error(err)
	}
	if err := msg.WriteTo(w); err != nil {
		t.Error(err)
	}
	if err := w.Flush(); err != nil {
		t.Error(err)
	}

	r := bufio.NewReader(&buf)
	msg2, err := ReadMessage(r)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(msg.Header, msg2.Header) {
		t.Error("message headers differ", msg2.Header, "expected:", msg.Header)
	}

	if msg2.Body != msg.Body {
		t.Errorf("Bodies differ:\n%q\nExpected:\n%q", msg2.Body, msg.Body)
	}

	msg3, err := ReadMessage(r)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(msg.Header, msg3.Header) {
		t.Error("message headers differ", msg3.Header, "expected:", msg.Header)
	}

	if msg3.Body != msg.Body {
		t.Errorf("Bodies differ:\n%q\nExpected:\n%q", msg3.Body, msg.Body)
	}

}
