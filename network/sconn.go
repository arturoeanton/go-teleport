package network

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/klauspost/compress/zstd"
)

const (
	MsgTypeControl      = 0x01
	MsgTypeUncompressed = 0x02
	MsgTypeCompressed   = 0x03
	MsgTypeAuth         = 0x04
)

var sharedKey = []byte("thisis32byteslongthisis32byteslo") // 32 bytes exactos
var zstdDecoder, _ = zstd.NewReader(nil)
var zstdEncoderPool = sync.Pool{
	New: func() interface{} {
		enc, _ := zstd.NewWriter(nil)
		return enc
	},
}

func init() {
	tmpKey := os.Getenv("SHARED_KEY")
	if tmpKey != "" {
		sharedKey = []byte(tmpKey)
	} else {
		log.Printf("Usando SHARED_KEY por defecto")
	}
	if len(sharedKey) != 32 {
		log.Fatal("sharedKey debe tener exactamente 32 bytes")
	}
}

type SecureConn struct {
	Conn net.Conn
}

type SecureFrame struct {
	Type    byte   // 0x01 = control, 0x02 = data, etc.
	Payload []byte // contenido que realmente querés transportar
}

func NewSecureConn(c net.Conn) *SecureConn {
	return &SecureConn{Conn: c}
}

func (sc *SecureConn) Write(p []byte) (int, error) {
	var compressedBuffer bytes.Buffer
	enc := zstdEncoderPool.Get().(*zstd.Encoder)
	enc.Reset(&compressedBuffer)
	_, _ = enc.Write(p)
	_ = enc.Close()
	zstdEncoderPool.Put(enc)
	var frame SecureFrame
	if compressedBuffer.Len() < len(p) {
		log.Printf("[SecureConn][WRITE] Usando tipo 0x03 (comprimido)")
		frame = SecureFrame{Type: MsgTypeCompressed, Payload: compressedBuffer.Bytes()}
		_, err := sc.WriteFrame(frame)
		if err != nil {
			return 0, err
		}
		return len(p), nil
	} else {
		log.Printf("[SecureConn][WRITE] Usando tipo 0x02 (sin comprimir)")
		frame = SecureFrame{Type: MsgTypeUncompressed, Payload: p}
		return sc.WriteFrame(frame)
	}

}

func (sc *SecureConn) WriteFrame(frame SecureFrame) (int, error) {
	plainBuf := new(bytes.Buffer)
	plainBuf.WriteByte(frame.Type)
	plainBuf.Write(frame.Payload)

	cipherText, err := encrypt(plainBuf.Bytes())
	if err != nil {
		return 0, err
	}

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(cipherText)))

	if _, err := sc.Conn.Write(lenBuf); err != nil {
		return 0, err
	}
	if _, err := sc.Conn.Write(cipherText); err != nil {
		return 0, err
	}
	return len(frame.Payload), nil
}

func (sc *SecureConn) Read(p []byte) (int, error) {
	var lenBuf [4]byte
	if _, err := io.ReadFull(sc.Conn, lenBuf[:]); err != nil {
		return 0, err
	}
	cipherLen := binary.BigEndian.Uint32(lenBuf[:])
	if cipherLen < 13 {
		return 0, errors.New("invalid cipher length")
	}

	encrypted := make([]byte, cipherLen)
	if _, err := io.ReadFull(sc.Conn, encrypted); err != nil {
		return 0, err
	}

	plain, err := decrypt(encrypted)
	if err != nil {
		return 0, err
	}
	if len(plain) < 1 {
		return 0, errors.New("invalid decrypted payload")
	}

	typeByte := plain[0]
	payload := plain[1:]

	switch typeByte {
	case MsgTypeUncompressed:
		copy(p, payload)
		return len(payload), nil
	case MsgTypeCompressed:
		decompressed, err := zstdDecoder.DecodeAll(payload, nil)
		if err != nil {
			return 0, err
		}
		copy(p, decompressed)
		return len(decompressed), nil
	case MsgTypeAuth:
		if string(payload) != os.Getenv("AUTH_TOKEN") {
			log.Println("[SecureConn][READ] AUTH_TOKEN Error [❌]")
			sc.Close()
		}
		log.Println("[SecureConn][READ] AUTH_TOKEN Success [✅]")
		return len(payload), nil

	default:
		return 0, errors.New("unknown message type")
	}
}

// Funciones delegadas de net.Conn
func (sc *SecureConn) LocalAddr() net.Addr                { return sc.Conn.LocalAddr() }
func (sc *SecureConn) RemoteAddr() net.Addr               { return sc.Conn.RemoteAddr() }
func (sc *SecureConn) SetDeadline(t time.Time) error      { return sc.Conn.SetDeadline(t) }
func (sc *SecureConn) SetReadDeadline(t time.Time) error  { return sc.Conn.SetReadDeadline(t) }
func (sc *SecureConn) SetWriteDeadline(t time.Time) error { return sc.Conn.SetWriteDeadline(t) }
func (sc *SecureConn) Close() error                       { return sc.Conn.Close() }

// TCP tuning (seguridad y performance)
func (sc *SecureConn) SetNoDelay(noDelay bool) error {
	if tcp, ok := sc.Conn.(*net.TCPConn); ok {
		return tcp.SetNoDelay(noDelay)
	}
	return errors.New("not a TCP connection")
}
func (sc *SecureConn) SetKeepAlive(v bool) error {
	if tcp, ok := sc.Conn.(*net.TCPConn); ok {
		return tcp.SetKeepAlive(v)
	}
	return errors.New("not a TCP connection")
}
func (sc *SecureConn) SetKeepAlivePeriod(d time.Duration) error {
	if tcp, ok := sc.Conn.(*net.TCPConn); ok {
		return tcp.SetKeepAlivePeriod(d)
	}
	return errors.New("not a TCP connection")
}
func (sc *SecureConn) SetLinger(sec int) error {
	if tcp, ok := sc.Conn.(*net.TCPConn); ok {
		return tcp.SetLinger(sec)
	}
	return errors.New("not a TCP connection")
}
func (sc *SecureConn) SetReadBuffer(sz int) error {
	if tcp, ok := sc.Conn.(*net.TCPConn); ok {
		return tcp.SetReadBuffer(sz)
	}
	return errors.New("not a TCP connection")
}
func (sc *SecureConn) SetWriteBuffer(sz int) error {
	if tcp, ok := sc.Conn.(*net.TCPConn); ok {
		return tcp.SetWriteBuffer(sz)
	}
	return errors.New("not a TCP connection")
}

// Encryption helpers
func encrypt(plain []byte) ([]byte, error) {
	block, err := aes.NewCipher(sharedKey)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, 12)
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	cipherText := aesgcm.Seal(nil, nonce, plain, nil)
	return append(nonce, cipherText...), nil
}

func decrypt(cipherText []byte) ([]byte, error) {
	if len(cipherText) < 12 {
		return nil, errors.New("cipherText too short")
	}
	block, err := aes.NewCipher(sharedKey)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := cipherText[:12]
	data := cipherText[12:]
	return aesgcm.Open(nil, nonce, data, nil)
}
