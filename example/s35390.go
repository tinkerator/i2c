// Program s34390 is an example that explores the properties of a
// real time clock chip. The data sheet for which is:
//
//	https://www.ablic.com/en/doc/datasheet/real_time_clock/S35390A_E.pdf
Package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math/bits"
	"time"

	"zappem.net/pub/io/i2c"
)

// sr1 etc are i2c device offset addresses for the various device
// functions.
const (
	sr1 = 0x30 + iota
	sr2
	rtd1
	rtd2
	int1
	int2
	clkc
	free
)

var (
	reset    = flag.Bool("reset", false, "force a reset")
	clock    = flag.Bool("clock", false, "track the time")
	set      = flag.Bool("set", false, "set the date, in UTC, from the system clock")
	duration = flag.Duration("watch", 3*time.Minute, "time to watch the clock for")
)

// toDeviceBCD converts the assumed 8-bit integer to the BCD format
// the device expects.
func toDeviceBCD(v int) byte {
	return bits.Reverse8(uint8(v/10)<<4 + uint8(v%10))
}

// fromDeviceBCD converts a device byte to an integer.
func fromDeviceBCD(u byte) int {
	v := bits.Reverse8(uint8(u))
	return int((v>>4)*10 + (v & 0xf))
}

// Now reads the device and unpacks its time value.
func Now() (time.Time, error) {
	d, err := i2c.NewConn(i2c.BusFile(1), rtd1, false, binary.LittleEndian)
	if err != nil {
		return time.Time{}, err
	}
	defer d.Close()
	buf := make([]byte, 7)
	if n, err := d.Read(buf); n != 7 || err != nil {
		return time.Time{}, fmt.Errorf("failed to read time n=%d vs. 7 bytes: %v", n, err)
	}
	return time.Date(2000+fromDeviceBCD(buf[0]), time.Month(fromDeviceBCD(buf[1])), fromDeviceBCD(buf[2]), fromDeviceBCD(buf[4] & ^byte(3)), fromDeviceBCD(buf[5]), fromDeviceBCD(buf[6]), 0, time.UTC), nil
}

// SetDate transcribes the UTC clock from the system and sets the
// clock managed by the s34390 device. The speed of light for this
// function to execute is the returned duration. Care is taken to
// align the clock to the system clock. The function can take some
// time to complete because of this alignment.
func SetDate(delay time.Duration) (time.Duration, error) {
	d, err := i2c.NewConn(i2c.BusFile(1), rtd1, false, binary.LittleEndian)
	if err != nil {
		return 0, err
	}
	defer d.Close()

	// From here to [**] below is expected to take delay duration.
	t1 := time.Now()

	goal := t1.Add(500*time.Millisecond + delay).Round(time.Second).In(time.UTC)
	t1 = goal.Add(-delay)
	for delta := goal.Sub(time.Now()); delta > 0; delta /= 2 {
		time.Sleep(delta)
	}
	t1 = time.Now()

	y := goal.Year() - 2000
	if y >= 100 || y < 0 {
		return 0, fmt.Errorf("time %q is not settable in this device", goal)
	}
	buf := make([]byte, 7)
	buf[0] = toDeviceBCD(y)
	buf[1] = toDeviceBCD(int(goal.Month()))
	buf[2] = toDeviceBCD(goal.Day())
	buf[3] = toDeviceBCD(int(goal.Weekday()))
	h := goal.Hour()
	pm := byte(2)
	if h < 12 {
		pm = 0
	}
	buf[4] = toDeviceBCD(h) | pm
	buf[5] = toDeviceBCD(goal.Minute())
	buf[6] = toDeviceBCD(goal.Second())
	d.Write(buf)

	t2 := time.Now()
	rt, err := Now()
	// [**] provided estimate was from here.
	t3 := time.Now()

	log.Printf("t1: %v", t1)
	log.Printf("go: %v", goal)
	log.Printf("t2: %v", t2)
	log.Printf("t3: %v", t3)
	log.Printf("rt: %v (%v)", rt, err)

	return (t3.Sub(t1) + t2.Sub(t1) + 1) / 2, err
}

func main() {
	flag.Parse()
	dump := func(addr, n uint) {
		d := make([]byte, n)
		c, err := i2c.NewConn(i2c.BusFile(1), addr, false, binary.LittleEndian)
		if err == nil {
			var m int
			m, err = c.Read(d)
			c.Close()
			if err == nil && uint(m) != n {
				err = i2c.ErrTruncated
			}
		}
		for i, v := range d {
			r := bits.Reverse8(uint8(v))
			log.Printf("read[%x;%d]: %02x %08b %08b %d (%v)", addr, i, v, v, r, r, err)
		}
	}

	s1, err := i2c.NewConn(i2c.BusFile(1), sr1, false, binary.LittleEndian)
	if err != nil {
		log.Fatalf("failed to open device sr1: %v", err)
	}
	defer s1.Close()

	r2, err := i2c.NewConn(i2c.BusFile(1), rtd2, false, binary.LittleEndian)
	if err != nil {
		log.Fatalf("failed to open device rtd2: %v", err)
	}
	defer r2.Close()

	if *reset {
		dump(sr1, 1)
		time.Sleep(time.Second / 2)
		s1.Write([]byte{0x80 + 0x40}) // Reset to be a 24 hour clock.
	}

	if *set {
		latency, err := SetDate(0)
		log.Printf("[0] latency=%v, err=%v", latency, err)
		latency, err = SetDate(latency)
		log.Printf("[1] latency=%v, err=%v", latency, err)
	}

	if *clock {
		var ot time.Time
		target := time.Now().Add(*duration).In(time.UTC)
		for time.Now().Before(target) {
			t, err := Now()
			if ot == t {
				continue
			}
			ot = t
			log.Printf("%v %v", t, err)
			time.Sleep(233 * time.Millisecond)
		}
		return
	}

	s2, err := i2c.NewConn(i2c.BusFile(1), sr2, false, binary.LittleEndian)
	if err != nil {
		log.Fatalf("failed to open device sr1: %v", err)
	}
	defer s2.Close()

	i1, err := i2c.NewConn(i2c.BusFile(1), int1, false, binary.LittleEndian)
	if err != nil {
		log.Fatalf("failed to open device sr1: %v", err)
	}
	defer i1.Close()

	dump(sr1, 1)
	dump(sr2, 1)
	dump(rtd1, 7)
	dump(rtd2, 3)
	dump(int1, 1)
	dump(int2, 1)
	dump(clkc, 1)
	dump(free, 1)
}
