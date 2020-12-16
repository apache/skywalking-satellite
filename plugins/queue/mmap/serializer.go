// Licensed to Apache Software Foundation (ASF) under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Apache Software Foundation (ASF) licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package mmap

import (
	"bytes"

	"encoding/gob"

	"github.com/apache/skywalking-satellite/internal/pkg/event"
)

// Decoder used in pop operation for reusing gob.Decoder and buf.
type Decoder struct {
	buf     *bytes.Buffer
	decoder *gob.Decoder
}

// Encoder used in push operation for reusing gob.Decoder and buf.
type Encoder struct {
	buf     *bytes.Buffer
	encoder *gob.Encoder
}

func NewDecoder() *Decoder {
	buf := new(bytes.Buffer)
	return &Decoder{
		buf:     buf,
		decoder: gob.NewDecoder(buf),
	}
}

func NewEncoder() *Encoder {
	buf := new(bytes.Buffer)
	return &Encoder{
		buf:     buf,
		encoder: gob.NewEncoder(buf),
	}
}

func (d *Decoder) deserialize(b []byte) (*event.Event, error) {
	defer d.buf.Reset()
	d.buf.Write(b)
	e := &event.Event{}
	err := d.decoder.Decode(e)
	if err != nil {
		return nil, err
	}
	return e, nil
}

func (e *Encoder) serialize(data *event.Event) ([]byte, error) {
	defer e.buf.Reset()
	err := e.encoder.Encode(data)
	if err != nil {
		return nil, err
	}
	return e.buf.Bytes(), nil
}
