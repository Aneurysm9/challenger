package vm

import (
	"fmt"
	"io/ioutil"

	stck "github.com/golang-collections/collections/stack"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrorHalt indicates that the machine has received a Halt instruction
	ErrorHalt = fmt.Errorf("The machine has halted")
	// ErrorUnknownInstruction indicates that the machine received an unknown instruction
	ErrorUnknownInstruction = fmt.Errorf("An unknown instruction was received")
)

const (
	maxSize uint16 = 1 << 15
)

// Machine represents an instance of the Synacor Challenge VM
type Machine struct {
	memory    []uint16
	registers [8]uint16
	stack     *stck.Stack
	ip        uint16
	Debug     bool
}

// NewMachine creates a new Machine instance
func NewMachine() *Machine {
	return &Machine{memory: make([]uint16, 1<<15), stack: stck.New(), Debug: false}
}

// LoadImage loads a machine memory image into a new machine
func LoadImage(fn string) (*Machine, error) {
	img, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	m := NewMachine()

	for i := 0; i < int(maxSize); i++ {
		if 2*i >= len(img) {
			break
		}
		m.memory[i] = uint16(img[2*i]) | (uint16(img[2*i+1]) << 8)
	}

	m.Debug = true
	log.SetLevel(log.DebugLevel)
	return m, nil
}

// Run starts a Machine
func (m *Machine) Run() error {
	for {
		if err := m.next(); err != nil {
			return err
		}
	}
}

func (m *Machine) next() error {
	switch m.memory[m.ip] {
	case 0:
		return ErrorHalt
	case 1:
		return m.set()
	case 2:
		return m.push()
	case 3:
		return m.pop()
	case 6:
		return m.jump()
	case 7:
		return m.jt()
	case 8:
		return m.jf()
	case 19:
		return m.out()
	case 21:
		return m.noop()
	}

	log.WithFields(log.Fields{
		"ip":    m.ip,
		"instr": m.memory[m.ip],
	}).Fatal("Unknown instruction")
	return ErrorUnknownInstruction
}

func (m *Machine) set() error {
	idx := m.memory[m.ip+1] % maxSize
	if idx > 7 {
		log.WithFields(log.Fields{
			"ip":  m.ip,
			"op":  "set",
			"idx": idx,
		}).Error("Attempt to set an invalid register")
		return fmt.Errorf("Attempt to set an invalid register")
	}

	val := m.getVal(m.ip + 2)

	if m.Debug {
		log.WithFields(log.Fields{
			"ip":  m.ip,
			"op":  "set",
			"idx": idx,
			"val": val,
		}).Debug("Setting register")
	}
	m.registers[idx] = val

	m.ip += 3
	return nil
}

func (m *Machine) push() error {
	val := m.memory[m.ip+1]
	if m.Debug {
		log.WithFields(log.Fields{
			"ip":  m.ip,
			"val": val,
		}).Debug("Pushing value onto stack")
	}

	m.stack.Push(val)
	m.ip += 2
	return nil
}

func (m *Machine) pop() error {
	dest := m.memory[m.ip+1]
	val := m.stack.Pop()
	if m.Debug {
		log.WithFields(log.Fields{
			"ip":   m.ip,
			"dest": dest,
			"val":  val,
		}).Debug("Popping value from stack")
	}

	m.setVal(dest, val.(uint16))
	m.ip += 2
	return nil
}

func (m *Machine) jump() error {
	dest := m.memory[m.ip+1]
	if m.Debug {
		log.WithFields(log.Fields{
			"ip":   m.ip,
			"dest": dest,
		}).Debug("Jumping")
	}

	m.ip = dest
	return nil
}

func (m *Machine) jt() error {
	val := m.getVal(m.ip + 1)
	dest := m.getVal(m.ip + 2)
	if m.Debug {
		log.WithFields(log.Fields{
			"ip":   m.ip,
			"dest": dest,
			"val":  val,
		}).Debug("Jumping if true")
	}

	if val != 0 {
		m.ip = dest
	} else {
		m.ip += 3
	}
	return nil
}

func (m *Machine) jf() error {
	val := m.getVal(m.ip + 1)
	dest := m.getVal(m.ip + 2)
	if m.Debug {
		log.WithFields(log.Fields{
			"ip":   m.ip,
			"dest": dest,
			"val":  val,
		}).Debug("Jumping if false")
	}

	if val == 0 {
		m.ip = dest
	} else {
		m.ip += 3
	}
	return nil
}

func (m *Machine) out() error {
	val := m.memory[m.ip+1]
	if m.Debug {
		log.WithFields(log.Fields{
			"ip":  m.ip,
			"op":  "out",
			"val": val,
		}).Debug("Printing character")
	}
	fmt.Printf("%c", val)
	m.ip += 2
	return nil
}

func (m *Machine) noop() error {
	if m.Debug {
		log.WithFields(log.Fields{
			"ip": m.ip,
			"op": "noop",
		}).Debug("Nothing to see here...")
	}
	m.ip++
	return nil
}

func (m *Machine) getVal(loc uint16) uint16 {
	val := m.memory[loc]
	if isReg(val) {
		return m.registers[val%maxSize]
	}
	return val
}

func (m *Machine) setVal(loc, val uint16) {
	if isReg(loc) {
		m.registers[loc%maxSize] = val
	} else {
		m.memory[loc] = val
	}
}

func isMem(v uint16) bool {
	return v < maxSize
}

func isReg(v uint16) bool {
	return v >= maxSize
}
