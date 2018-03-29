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
	maxSize int = 1 << 15
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

	for i := 0; i < maxSize; i++ {
		if 2*i >= len(img) {
			break
		}
		m.memory[i] = uint16(img[2*i]) | (uint16(img[2*i+1]) << 8)
	}

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
	case 19:
		return m.out()
	case 21:
		return m.noop()
	}

	return ErrorUnknownInstruction
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
