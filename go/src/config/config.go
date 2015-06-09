package config

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type VersionedConfig interface {
	GetVersion() int /* Return the version */
	IncVersion()     /* Increment version */
}

func readConfig(file io.Reader, data VersionedConfig) (err error) {
	decoder := json.NewDecoder(file)
	err = decoder.Decode(data)
	return
}

func ReadConfig(filename string, data VersionedConfig) (err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()
	err = readConfig(file, data)
	
	return
}

type VersionMismatch struct {
	ondisk, memory int
}

func (e VersionMismatch) Error() string {
	return fmt.Sprintf("Version mismatch - on disk %v, attempting %v\n",
		e.ondisk, e.memory)
}

func EncodeObject(w io.Writer, data VersionedConfig) (err error) {
	encoder := json.NewEncoder(w)
	err = encoder.Encode(data)
	return
}

func DecodeObject(r io.Reader, data VersionedConfig) (err error) {
	decoder := json.NewDecoder(r)
	err = decoder.Decode(&data)
	return
}

func WriteConfig(filename string, data VersionedConfig) (err error) {
	var ondisk int
	ofile, err := os.OpenFile(filename, os.O_RDONLY, os.ModeExclusive)
	if err != nil {
		ondisk = 0
	} else {
		defer ofile.Close()
		var tmpdata map[string]interface{}
		decoder := json.NewDecoder(ofile)
		err = decoder.Decode(&tmpdata)

		if err != nil {
			return
		}
		if tmpdata["Version"] != nil {
			ondisk = int(tmpdata["Version"].(float64))
		}
	}

	if ondisk != 0 && ondisk > data.GetVersion() {
		return VersionMismatch{ondisk,
			data.GetVersion()}
	}
	file, err := ioutil.TempFile("", "config")
	if err != nil {
		return
	}
	defer file.Close()
	err = EncodeObject(file, data)
	if err != nil {
		return
	}
	err = os.Rename(file.Name(), filename)
	return
}
