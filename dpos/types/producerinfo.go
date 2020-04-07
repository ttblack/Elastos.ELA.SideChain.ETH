// Copyright (c) 2017-2019 The Elastos Foundation
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.
//

package types

import (
	"bytes"
	"io"
)

const ProducerInfoVersion byte = 0x00

type ProducerInfo struct {
	OwnerPublicKey []byte
	NodePublicKey  []byte
	NickName       string
	Url            string
	Location       uint64
	NetAddress     string
	Signature      []byte
}

func (a *ProducerInfo) Data(version byte) []byte {
	buf := new(bytes.Buffer)
	if err := a.Serialize(buf, version); err != nil {
		return []byte{0}
	}
	return buf.Bytes()
}

func (a *ProducerInfo) Serialize(w io.Writer, version byte) error {
	return nil
}

func (a *ProducerInfo) SerializeUnsigned(w io.Writer, version byte) error {
	return nil
}

func (a *ProducerInfo) Deserialize(r io.Reader, version byte) error {
	return nil
}

func (a *ProducerInfo) DeserializeUnsigned(r io.Reader, version byte) error {
	return nil
}
