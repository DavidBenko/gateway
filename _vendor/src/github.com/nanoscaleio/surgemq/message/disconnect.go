// Copyright (c) 2014 The SurgeMQ Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package message

import "fmt"

// The DISCONNECT Packet is the final Control Packet sent from the Client to the Server.
// It indicates that the Client is disconnecting cleanly.
type DisconnectMessage struct {
	header
}

var _ Message = (*DisconnectMessage)(nil)

// NewDisconnectMessage creates a new DISCONNECT message.
func NewDisconnectMessage() *DisconnectMessage {
	msg := &DisconnectMessage{}
	msg.SetType(DISCONNECT)

	return msg
}

func (this *DisconnectMessage) Decode(src []byte) (int, error) {
	return this.header.decode(src)
}

func (this *DisconnectMessage) Encode(dst []byte) (int, error) {
	if !this.Dirty {
		if len(dst) < len(this.Dbuf) {
			return 0, fmt.Errorf("disconnect/Encode: Insufficient buffer size. Expecting %d, got %d.", len(this.Dbuf), len(dst))
		}

		return copy(dst, this.Dbuf), nil
	}

	return this.header.encode(dst)
}
