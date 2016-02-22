// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"crypto"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"sync"
)

// clientVersion is the fixed identification string that the client will use.
var clientVersion = []byte("SSH-2.0-Go\r\n")

// ClientConn represents the client side of an SSH connection.
type ClientConn struct {
	*transport
	config *ClientConfig
	chanlist
}

// Client returns a new SSH client connection using c as the underlying transport.
func Client(c net.Conn, config *ClientConfig) (*ClientConn, error) {
	conn := &ClientConn{
		transport: newTransport(c, config.rand()),
		config:    config,
	}
	if err := conn.handshake(); err != nil {
		conn.Close()
		return nil, err
	}
	go conn.mainLoop()
	return conn, nil
}

// handshake performs the client side key exchange. See RFC 4253 Section 7.
func (c *ClientConn) handshake() error {
	var magics handshakeMagics

	if _, err := c.Write(clientVersion); err != nil {
		return err
	}
	if err := c.Flush(); err != nil {
		return err
	}
	magics.clientVersion = clientVersion[:len(clientVersion)-2]

	// read remote server version
	version, err := readVersion(c)
	if err != nil {
		return err
	}
	magics.serverVersion = version
	clientKexInit := kexInitMsg{
		KexAlgos:                supportedKexAlgos,
		ServerHostKeyAlgos:      supportedHostKeyAlgos,
		CiphersClientServer:     c.config.Crypto.ciphers(),
		CiphersServerClient:     c.config.Crypto.ciphers(),
		MACsClientServer:        supportedMACs,
		MACsServerClient:        supportedMACs,
		CompressionClientServer: supportedCompressions,
		CompressionServerClient: supportedCompressions,
	}
	kexInitPacket := marshal(msgKexInit, clientKexInit)
	magics.clientKexInit = kexInitPacket

	if err := c.writePacket(kexInitPacket); err != nil {
		return err
	}
	packet, err := c.readPacket()
	if err != nil {
		return err
	}

	magics.serverKexInit = packet

	var serverKexInit kexInitMsg
	if err = unmarshal(&serverKexInit, packet, msgKexInit); err != nil {
		return err
	}

	kexAlgo, hostKeyAlgo, ok := findAgreedAlgorithms(c.transport, &clientKexInit, &serverKexInit)
	if !ok {
		return errors.New("ssh: no common algorithms")
	}

	if serverKexInit.FirstKexFollows && kexAlgo != serverKexInit.KexAlgos[0] {
		// The server sent a Kex message for the wrong algorithm,
		// which we have to ignore.
		if _, err := c.readPacket(); err != nil {
			return err
		}
	}

	var H, K []byte
	var hashFunc crypto.Hash
	switch kexAlgo {
	case kexAlgoDH14SHA1:
		hashFunc = crypto.SHA1
		dhGroup14Once.Do(initDHGroup14)
		H, K, err = c.kexDH(dhGroup14, hashFunc, &magics, hostKeyAlgo)
	default:
		err = fmt.Errorf("ssh: unexpected key exchange algorithm %v", kexAlgo)
	}
	if err != nil {
		return err
	}

	if err = c.writePacket([]byte{msgNewKeys}); err != nil {
		return err
	}
	if err = c.transport.writer.setupKeys(clientKeys, K, H, H, hashFunc); err != nil {
		return err
	}
	if packet, err = c.readPacket(); err != nil {
		return err
	}
	if packet[0] != msgNewKeys {
		return UnexpectedMessageError{msgNewKeys, packet[0]}
	}
	if err := c.transport.reader.setupKeys(serverKeys, K, H, H, hashFunc); err != nil {
		return err
	}
	return c.authenticate(H)
}

// kexDH performs Diffie-Hellman key agreement on a ClientConn. The
// returned values are given the same names as in RFC 4253, section 8.
func (c *ClientConn) kexDH(group *dhGroup, hashFunc crypto.Hash, magics *handshakeMagics, hostKeyAlgo string) ([]byte, []byte, error) {
	x, err := rand.Int(c.config.rand(), group.p)
	if err != nil {
		return nil, nil, err
	}
	X := new(big.Int).Exp(group.g, x, group.p)
	kexDHInit := kexDHInitMsg{
		X: X,
	}
	if err := c.writePacket(marshal(msgKexDHInit, kexDHInit)); err != nil {
		return nil, nil, err
	}

	packet, err := c.readPacket()
	if err != nil {
		return nil, nil, err
	}

	var kexDHReply = new(kexDHReplyMsg)
	if err = unmarshal(kexDHReply, packet, msgKexDHReply); err != nil {
		return nil, nil, err
	}

	if kexDHReply.Y.Sign() == 0 || kexDHReply.Y.Cmp(group.p) >= 0 {
		return nil, nil, errors.New("server DH parameter out of bounds")
	}

	kInt := new(big.Int).Exp(kexDHReply.Y, x, group.p)
	h := hashFunc.New()
	writeString(h, magics.clientVersion)
	writeString(h, magics.serverVersion)
	writeString(h, magics.clientKexInit)
	writeString(h, magics.serverKexInit)
	writeString(h, kexDHReply.HostKey)
	writeInt(h, X)
	writeInt(h, kexDHReply.Y)
	K := make([]byte, intLength(kInt))
	marshalInt(K, kInt)
	h.Write(K)

	H := h.Sum(nil)

	return H, K, nil
}

// mainLoop reads incoming messages and routes channel messages
// to their respective ClientChans.
func (c *ClientConn) mainLoop() {
	// TODO(dfc) signal the underlying close to all channels
	defer c.Close()
	for {
		packet, err := c.readPacket()
		if err != nil {
			break
		}
		// TODO(dfc) A note on blocking channel use.
		// The msg, win, data and dataExt channels of a clientChan can
		// cause this loop to block indefinately if the consumer does
		// not service them.
		switch packet[0] {
		case msgChannelData:
			if len(packet) < 9 {
				// malformed data packet
				break
			}
			peersId := uint32(packet[1])<<24 | uint32(packet[2])<<16 | uint32(packet[3])<<8 | uint32(packet[4])
			if length := int(packet[5])<<24 | int(packet[6])<<16 | int(packet[7])<<8 | int(packet[8]); length > 0 {
				packet = packet[9:]
				c.getChan(peersId).stdout.handleData(packet[:length])
			}
		case msgChannelExtendedData:
			if len(packet) < 13 {
				// malformed data packet
				break
			}
			peersId := uint32(packet[1])<<24 | uint32(packet[2])<<16 | uint32(packet[3])<<8 | uint32(packet[4])
			datatype := uint32(packet[5])<<24 | uint32(packet[6])<<16 | uint32(packet[7])<<8 | uint32(packet[8])
			if length := int(packet[9])<<24 | int(packet[10])<<16 | int(packet[11])<<8 | int(packet[12]); length > 0 {
				packet = packet[13:]
				// RFC 4254 5.2 defines data_type_code 1 to be data destined
				// for stderr on interactive sessions. Other data types are
				// silently discarded.
				if datatype == 1 {
					c.getChan(peersId).stderr.handleData(packet[:length])
				}
			}
		default:
			switch msg := decode(packet).(type) {
			case *channelOpenMsg:
				c.getChan(msg.PeersId).msg <- msg
			case *channelOpenConfirmMsg:
				c.getChan(msg.PeersId).msg <- msg
			case *channelOpenFailureMsg:
				c.getChan(msg.PeersId).msg <- msg
			case *channelCloseMsg:
				ch := c.getChan(msg.PeersId)
				ch.theyClosed = true
				close(ch.stdin.win)
				ch.stdout.eof()
				ch.stderr.eof()
				close(ch.msg)
				if !ch.weClosed {
					ch.weClosed = true
					ch.sendClose()
				}
				c.chanlist.remove(msg.PeersId)
			case *channelEOFMsg:
				ch := c.getChan(msg.PeersId)
				ch.stdout.eof()
				// RFC 4254 is mute on how EOF affects dataExt messages but
				// it is logical to signal EOF at the same time.
				ch.stderr.eof()
			case *channelRequestSuccessMsg:
				c.getChan(msg.PeersId).msg <- msg
			case *channelRequestFailureMsg:
				c.getChan(msg.PeersId).msg <- msg
			case *channelRequestMsg:
				c.getChan(msg.PeersId).msg <- msg
			case *windowAdjustMsg:
				c.getChan(msg.PeersId).stdin.win <- int(msg.AdditionalBytes)
			case *disconnectMsg:
				break
			default:
				fmt.Printf("mainLoop: unhandled message %T: %v\n", msg, msg)
			}
		}
	}
}

// Dial connects to the given network address using net.Dial and
// then initiates a SSH handshake, returning the resulting client connection.
func Dial(network, addr string, config *ClientConfig) (*ClientConn, error) {
	conn, err := net.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	return Client(conn, config)
}

// A ClientConfig structure is used to configure a ClientConn. After one has
// been passed to an SSH function it must not be modified.
type ClientConfig struct {
	// Rand provides the source of entropy for key exchange. If Rand is
	// nil, the cryptographic random reader in package crypto/rand will
	// be used.
	Rand io.Reader

	// The username to authenticate.
	User string

	// A slice of ClientAuth methods. Only the first instance
	// of a particular RFC 4252 method will be used during authentication.
	Auth []ClientAuth

	// Cryptographic-related configuration.
	Crypto CryptoConfig
}

func (c *ClientConfig) rand() io.Reader {
	if c.Rand == nil {
		return rand.Reader
	}
	return c.Rand
}

// A clientChan represents a single RFC 4254 channel that is multiplexed
// over a single SSH connection.
type clientChan struct {
	packetWriter
	id, peersId uint32
	stdin       *chanWriter      // receives window adjustments
	stdout      *chanReader      // receives the payload of channelData messages
	stderr      *chanReader      // receives the payload of channelExtendedData messages
	msg         chan interface{} // incoming messages
	theyClosed  bool             // indicates the close msg has been received from the remote side
	weClosed    bool             // incidates the close msg has been sent from our side
}

// newClientChan returns a partially constructed *clientChan
// using the local id provided. To be usable clientChan.peersId
// needs to be assigned once known.
func newClientChan(t *transport, id uint32) *clientChan {
	c := &clientChan{
		packetWriter: t,
		id:           id,
		msg:          make(chan interface{}, 16),
	}
	c.stdin = &chanWriter{
		win:        make(chan int, 16),
		clientChan: c,
	}
	c.stdout = &chanReader{
		data:       make(chan []byte, 16),
		clientChan: c,
	}
	c.stderr = &chanReader{
		data:       make(chan []byte, 16),
		clientChan: c,
	}
	return c
}

// waitForChannelOpenResponse, if successful, fills out
// the peerId and records any initial window advertisement.
func (c *clientChan) waitForChannelOpenResponse() error {
	switch msg := (<-c.msg).(type) {
	case *channelOpenConfirmMsg:
		// fixup peersId field
		c.peersId = msg.MyId
		c.stdin.win <- int(msg.MyWindow)
		return nil
	case *channelOpenFailureMsg:
		return errors.New(safeString(msg.Message))
	}
	return errors.New("unexpected packet")
}

// sendEOF sends EOF to the server. RFC 4254 Section 5.3
func (c *clientChan) sendEOF() error {
	return c.writePacket(marshal(msgChannelEOF, channelEOFMsg{
		PeersId: c.peersId,
	}))
}

// sendClose signals the intent to close the channel.
func (c *clientChan) sendClose() error {
	return c.writePacket(marshal(msgChannelClose, channelCloseMsg{
		PeersId: c.peersId,
	}))
}

// Close closes the channel. This does not close the underlying connection.
func (c *clientChan) Close() error {
	if !c.weClosed {
		c.weClosed = true
		return c.sendClose()
	}
	return nil
}

// Thread safe channel list.
type chanlist struct {
	// protects concurrent access to chans
	sync.Mutex
	// chans are indexed by the local id of the channel, clientChan.id.
	// The PeersId value of messages received by ClientConn.mainLoop is
	// used to locate the right local clientChan in this slice.
	chans []*clientChan
}

// Allocate a new ClientChan with the next avail local id.
func (c *chanlist) newChan(t *transport) *clientChan {
	c.Lock()
	defer c.Unlock()
	for i := range c.chans {
		if c.chans[i] == nil {
			ch := newClientChan(t, uint32(i))
			c.chans[i] = ch
			return ch
		}
	}
	i := len(c.chans)
	ch := newClientChan(t, uint32(i))
	c.chans = append(c.chans, ch)
	return ch
}

func (c *chanlist) getChan(id uint32) *clientChan {
	c.Lock()
	defer c.Unlock()
	return c.chans[int(id)]
}

func (c *chanlist) remove(id uint32) {
	c.Lock()
	defer c.Unlock()
	c.chans[int(id)] = nil
}

// A chanWriter represents the stdin of a remote process.
type chanWriter struct {
	win        chan int    // receives window adjustments
	rwin       int         // current rwin size
	clientChan *clientChan // the channel backing this writer
}

// Write writes data to the remote process's standard input.
func (w *chanWriter) Write(data []byte) (written int, err error) {
	for len(data) > 0 {
		for w.rwin < 1 {
			win, ok := <-w.win
			if !ok {
				return 0, io.EOF
			}
			w.rwin += win
		}
		n := min(len(data), w.rwin)
		peersId := w.clientChan.peersId
		packet := []byte{
			msgChannelData,
			byte(peersId >> 24), byte(peersId >> 16), byte(peersId >> 8), byte(peersId),
			byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n),
		}
		if err = w.clientChan.writePacket(append(packet, data[:n]...)); err != nil {
			break
		}
		data = data[n:]
		w.rwin -= n
		written += n
	}
	return
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (w *chanWriter) Close() error {
	return w.clientChan.sendEOF()
}

// A chanReader represents stdout or stderr of a remote process.
type chanReader struct {
	// TODO(dfc) a fixed size channel may not be the right data structure.
	// If writes to this channel block, they will block mainLoop, making
	// it unable to receive new messages from the remote side.
	data       chan []byte // receives data from remote
	dataClosed bool        // protects data from being closed twice
	clientChan *clientChan // the channel backing this reader
	buf        []byte
}

// eof signals to the consumer that there is no more data to be received.
func (r *chanReader) eof() {
	if !r.dataClosed {
		r.dataClosed = true
		close(r.data)
	}
}

// handleData sends buf to the reader's consumer. If r.data is closed
// the data will be silently discarded
func (r *chanReader) handleData(buf []byte) {
	if !r.dataClosed {
		r.data <- buf
	}
}

// Read reads data from the remote process's stdout or stderr.
func (r *chanReader) Read(data []byte) (int, error) {
	var ok bool
	for {
		if len(r.buf) > 0 {
			n := copy(data, r.buf)
			r.buf = r.buf[n:]
			msg := windowAdjustMsg{
				PeersId:         r.clientChan.peersId,
				AdditionalBytes: uint32(n),
			}
			return n, r.clientChan.writePacket(marshal(msgChannelWindowAdjust, msg))
		}
		r.buf, ok = <-r.data
		if !ok {
			return 0, io.EOF
		}
	}
	panic("unreachable")
}
