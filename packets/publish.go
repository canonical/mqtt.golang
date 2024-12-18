/*
 * Copyright (c) 2024 Contributors to the Eclipse Foundation
 *
 *  All rights reserved. This program and the accompanying materials
 *  are made available under the terms of the Eclipse Public License v2.0
 *  and Eclipse Distribution License v1.0 which accompany this distribution.
 *
 * The Eclipse Public License is available at
 *    https://www.eclipse.org/legal/epl-2.0/
 *  and the Eclipse Distribution License is available at
 *    http://www.eclipse.org/org/documents/edl-v10.php.
 *
 *  SPDX-License-Identifier: EPL-2.0 OR BSD-3-Clause
 */

package packets

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
)

// Publish is the Variable Header definition for a publish control packet
type Publish struct {
	Payload    []byte
	Topic      string
	Properties *Properties
	PacketID   uint16
	QoS        byte
	Duplicate  bool
	Retain     bool
}

func (p *Publish) String() string {
	return fmt.Sprintf("PUBLISH: PacketID:%d QOS:%d Topic:%s Duplicate:%t Retain:%t Payload:\n%s\nProperties\n%s", p.PacketID, p.QoS, p.Topic, p.Duplicate, p.Retain, string(p.Payload), p.Properties)
}

// SetIdentifier sets the packet identifier
func (p *Publish) SetIdentifier(packetID uint16) {
	p.PacketID = packetID
}

// Type returns the current packet type
func (s *Publish) Type() byte {
	return PUBLISH
}

// Unpack is the implementation of the interface required function for a packet
func (p *Publish) Unpack(r *bytes.Buffer, protocolVersion *byte) error {
	var err error
	p.Topic, err = readString(r)
	if err != nil {
		return err
	}
	if p.QoS > 0 {
		p.PacketID, err = readUint16(r)
		if err != nil {
			return err
		}
	}

	if *protocolVersion == MQTT_5 {
		err = p.Properties.Unpack(r, PUBLISH)
		if err != nil {
			return err
		}
	}

	p.Payload, err = ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	return nil
}

// Buffers is the implementation of the interface required function for a packet
func (p *Publish) Buffers() net.Buffers {
	var b bytes.Buffer
	writeString(p.Topic, &b)
	if p.QoS > 0 {
		_ = writeUint16(p.PacketID, &b)
	}
	idvp := p.Properties.Pack(PUBLISH)
	encodeVBIdirect(len(idvp), &b)
	return net.Buffers{b.Bytes(), idvp, p.Payload}
}

// WriteTo is the implementation of the interface required function for a packet
func (p *Publish) WriteTo(w io.Writer) (int64, error) {
	return p.ToControlPacket().WriteTo(w)
}

// ToControlPacket returns the packet as a ControlPacket
func (p *Publish) ToControlPacket() *ControlPacket {
	f := p.QoS << 1
	if p.Duplicate {
		f |= 1 << 3
	}
	if p.Retain {
		f |= 1
	}

	return &ControlPacket{FixedHeader: FixedHeader{Type: PUBLISH, Flags: f}, Content: p}
}
