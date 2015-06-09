package util

import (
	"config"
	"errors"
	"fmt"
	"math/rand"
)
const (
	Disabled = false
	Enabled = true
)
type failpoint struct {
	// Name - of the failpoint. Must be unique for a given context
	Name string
	// Enabled - dictates if the failpoint should be enabled
	Enabled bool
	// Probability - If enabled what is the chance that it will fail.
	// Set to 100 for always fail
	Probability int
	// Err - Error message used to create an error object when this 
	// failpoint is fired
	Err string
	Version int
}

func (this *FailPointMap) IncVersion() {

}

func (this FailPointMap) GetVersion() int {
	return 1
}

type FailPointMap map[string]failpoint

// NewFailPointMap returns a map that is backed up to the
// file fileName
func NewFailPointMap(fileName string) (fpmap FailPointMap) {
	fpmap = make(FailPointMap)
	config.ReadConfig(fileName, &fpmap)
	return
}

func (this *FailPointMap) Save(fileName string) (err error) {
	err = config.WriteConfig(fileName, this)
	return
}

func (this *FailPointMap) AddFailpoint(name string, enabled bool, probability int, err string) {
	(*this)[name] = failpoint{name, enabled, probability, err, 1}
}

func (this *FailPointMap) UpdateFailpoint(name string, enabled bool, probability int, msg string) {
	fp := (*this)[name]
	fp.Enabled = enabled
	fp.Probability = probability
	fp.Err = msg
	(*this)[name] = fp
}

func (this *FailPointMap) DeleteFailpoint(name string) {
	delete(*this, name)
}

func (this *FailPointMap) Failpoint(name string, arg interface {}) (err error) {
	err = nil
	fp := (*this)[name]
	if fp.Enabled {
		r := rand.Intn(100)
		if r < fp.Probability {
			err = errors.New(fmt.Sprintf("%s: %+v", fp.Err, arg))
		}
	}
	return
}
