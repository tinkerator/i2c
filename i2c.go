// Package i2c abstracts use of the Linux kernel's i2c/smbus device
// drivers.
package i2c // zappem.net/pub/io/i2c

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"
)

// RETRIES etc are from /usr/include/linux/i2c-dev.h
const (
	RETRIES     = 0x0701
	TIMEOUT     = 0x0702
	SLAVE       = 0x0703
	SLAVE_FORCE = 0x0706
	TENBIT      = 0x0704
	FUNCS       = 0x0705
	RDWR        = 0x0707
	PEC         = 0x0708
	SMBUS       = 0x0720
)

// Conn holds an open connection to an i2c device.
type Conn struct {
	bus    string
	addr   uint
	mu     sync.Mutex
	f      *os.File
	endian binary.ByteOrder
}

var (
	ErrInvalid   = errors.New("invalid connection")
	ErrClosed    = errors.New("connection closed")
	ErrTruncated = errors.New("truncated transaction")
)

// ioctl performs an ioctl on the open connection.
func (c *Conn) ioctl(cmd uintptr, arg uintptr) error {
	if c == nil {
		return ErrInvalid
	}
	if c.f == nil {
		return ErrClosed
	}
	sc, err := c.f.SyscallConn()
	if err != nil {
		return err
	}
	sc.Control(func(fd uintptr) {
		_, _, eno := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, arg)
		if eno != 0 {
			err = eno
		}
	})
	return err
}

// Close shuts down the open connection.
func (c *Conn) Close() error {
	if c == nil {
		return ErrInvalid
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.f == nil {
		return ErrClosed
	}
	err := c.f.Close()
	c.f = nil
	return err
}

// BusFile is a convenience file for locating the bus numbered device
// file.
func BusFile(n uint) string {
	return fmt.Sprintf("/dev/i2c-%d", n)
}

// NewConn establishes a new connection to an addressed
// device. Whether or not the device uses 10-bit addressing and which
// endianness it is are device specific considerations.
func NewConn(bus string, addr uint, tenBit bool, endian binary.ByteOrder) (*Conn, error) {
	f, err := os.OpenFile(bus, os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	c := &Conn{bus: bus, addr: addr, f: f}
	if tenBit {
		err = c.ioctl(TENBIT, 1)
	} else {
		err = c.ioctl(TENBIT, 0)
	}
	if err == nil {
		err = c.ioctl(SLAVE, uintptr(addr))
	}
	if err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

// read reads up to data bytes from the open connection.
func (c *Conn) Read(data []byte) (int, error) {
	if c == nil {
		return 0, ErrInvalid
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.f == nil {
		return 0, ErrClosed
	}
	return c.f.Read(data)
}

// write writes data bytes to the open connection.
func (c *Conn) Write(data []byte) (int, error) {
	if c == nil {
		return 0, ErrInvalid
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.f == nil {
		return 0, ErrClosed
	}
	return c.f.Write(data)
}

// ReadUint16 reads a uint16 value from an open connection.
func (c *Conn) ReadUint16() (uint16, error) {
	d := make([]byte, 2)
	if n, err := c.Read(d); err != nil {
		return 0, err
	} else if n != len(d) {
		return 0, ErrTruncated
	}
	return c.endian.Uint16(d), nil
}

// WriteUint16 writes a uint16 value to an open connection.
func (c *Conn) WriteUint16(val uint16) error {
	if c == nil {
		return ErrInvalid
	}
	d := make([]byte, 2)
	c.endian.PutUint16(d, val)
	if n, err := c.Write(d); err != nil {
		return err
	} else if n != len(d) {
		return ErrTruncated
	}
	return nil
}

// ReadUint32 reads a uint32 value from an open connection.
func (c *Conn) ReadUint32() (uint32, error) {
	d := make([]byte, 4)
	if n, err := c.Read(d); err != nil {
		return 0, err
	} else if n != len(d) {
		return 0, ErrTruncated
	}
	return c.endian.Uint32(d), nil
}

// WriteUint32 writes a uint32 value to an open connection.
func (c *Conn) WriteUint32(val uint32) error {
	if c == nil {
		return ErrInvalid
	}
	d := make([]byte, 4)
	c.endian.PutUint32(d, val)
	if n, err := c.Write(d); err != nil {
		return err
	} else if n != len(d) {
		return ErrTruncated
	}
	return nil
}

// ReadUint64 reads a uint64 value from an open connection.
func (c *Conn) ReadUint64() (uint64, error) {
	d := make([]byte, 8)
	if n, err := c.Read(d); err != nil {
		return 0, err
	} else if n != len(d) {
		return 0, ErrTruncated
	}
	return c.endian.Uint64(d), nil
}

// WriteUint64 writes a uint64 value to an open connection.
func (c *Conn) WriteUint64(val uint64) error {
	if c == nil {
		return ErrInvalid
	}
	d := make([]byte, 8)
	c.endian.PutUint64(d, val)
	if n, err := c.Write(d); err != nil {
		return err
	} else if n != len(d) {
		return ErrTruncated
	}
	return nil
}
