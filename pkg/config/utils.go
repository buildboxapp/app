package config

import (
	"path/filepath"
	"strconv"
	"time"
)

const sep = string(filepath.Separator)

//Float32 custom duration for toml configs
type Float struct {
	float64
	Value float64
}

//UnmarshalText method satisfying toml unmarshal interface
func (d *Float) UnmarshalText(text []byte) error {
	var err error
	i, err := strconv.ParseFloat(string(text), 10)
	d.Value = i
	return err
}

//Duration custom duration for toml configs
type Duration struct {
	time.Duration
	Value time.Duration
}

//UnmarshalText method satisfying toml unmarshal interface
func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	t := string(text)
	// если получили только цифру - добавляем секунды (по-умолчанию)
	if len(t) != 0 {
		lastStr := t[len(t)-1:]
		if lastStr != "h" && lastStr != "m" && lastStr != "s" {
			t = t + "m"
		}
	}
	d.Value, err = time.ParseDuration(t)
	return err
}

//Duration custom duration for toml configs
type Int struct {
	int
	Value int
}

//UnmarshalText method satisfying toml unmarshal interface
func (d *Int) UnmarshalText(text []byte) error {
	var err error
	i, err := strconv.Atoi(string(text))
	d.Value = i
	return err
}
