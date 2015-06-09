package config

import (
	"bytes"
	"os"
	"testing"
)

type MyCfg struct {
	Version int
	Bla     string
	Sully   map[string]int
}

func (cfg MyCfg) GetVersion() int {
	return cfg.Version
}
func (cfg *MyCfg) IncVersion() {
	cfg.Version++
}

func TestConfig(t *testing.T) {
	configFile := "config.json"
	// Remove file before and after test
	os.Remove(configFile)
	defer os.Remove(configFile)

	/* Test Empty Config */
	cfg := MyCfg{0, "asdf", map[string]int{"kk": 2}}
	cfg.IncVersion()
	err := WriteConfig(configFile, &cfg)
	if err != nil {
		t.Error("Failed to WriteConfig\n")
	}

	err = ReadConfig(configFile, &cfg)
	if err != nil {
		t.Error("Failed to ReadConfig\n")
	}

	if cfg.GetVersion() != 1 {
		t.Error("Failed to Match Version\n")
	}

	/* Test version mismatch */
	cfg.Version = 0
	err = WriteConfig(configFile, &cfg)
	if err == nil {
		t.Error("Failed to get error for lower rev write\n")
	}

	/* Test Correct Version write */
	cfg.Version = 1
	cfg.IncVersion()
	err = WriteConfig("config.json", &cfg)
	if err != nil {
		t.Error("Failed to WriteConfig uprev\n")
	}

	err = ReadConfig(configFile, &cfg)
	if err != nil {
		t.Error("Failed to ReadConfig uprev\n")
	}

	if cfg.GetVersion() != 2 {
		t.Errorf("Failed to Match Version uprev: %d\n", cfg.GetVersion())
	}
}

type EncodeDecodeTest struct {
	StringValue string
	BoolValue bool
	IntValue int
	Version int
}

func (this EncodeDecodeTest) GetVersion () (int) {
	return this.Version
}

func (this EncodeDecodeTest) IncVersion () {
	this.Version++
}

func TestEncodeDecode(t *testing.T) {
	edt := EncodeDecodeTest{"Bla", true, 10, 1}
	var s string
	buf := bytes.NewBufferString(s)
	err := EncodeObject(buf, edt)
	if err != nil {
		t.Errorf("Failed to encode %+v", err)
	}
	s1 := buf.String()
	bufr := bytes.NewBufferString(s1)
	var edtr EncodeDecodeTest
	err = DecodeObject(bufr, &edtr)
	if err != nil {
		t.Errorf("Failed to decode %+v", err)
	}

	if edt != edtr {
		t.Errorf("%+v and %+v did not match", edt, edtr)
	}
}