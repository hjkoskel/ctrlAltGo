/*
 */
package initializing

import (
	"fmt"
	"os"
	"strings"
)

type KeyString struct {
	Key   string
	Value string
}

type KeyStringArr []KeyString

func ParseEnv(s string) (KeyString, error) {
	fields := strings.Split(s, "=")
	if len(fields) < 2 {
		return KeyString{}, fmt.Errorf("invalid environment variable %s, have fields %#v", s, fields)
	}
	return KeyString{Key: fields[0], Value: strings.Join(fields[1:], "=")}, nil
}

func GetEnvs() (KeyStringArr, error) {
	envs := os.Environ()
	result := make([]KeyString, len(envs))
	var err error
	for i, s := range envs {
		result[i], err = ParseEnv(s)
		if err != nil {
			return result, err
		}
	}
	return result, nil
}

func (a KeyStringArr) String() string {
	var sb strings.Builder
	for _, item := range a {
		sb.WriteString(fmt.Sprintf("%s=%s\n", item.Key, item.Value))
	}
	return sb.String()
}

func ParseEnvs(s string) (KeyStringArr, error) { //Reverse what String method does
	result := []KeyString{}
	rows := strings.Split(s, "\n")
	for _, row := range rows {
		if strings.TrimSpace(row) == "" {
			continue
		}
		entry, errParse := ParseEnv(row)
		if errParse != nil {
			return result, errParse
		}
		result = append(result, entry)
	}
	return result, nil
}

func (p *KeyString) SetEnv() error {
	return os.Setenv(p.Key, p.Value)
}

func (p *KeyStringArr) SetEnvs() error {
	for _, item := range *p {
		err := item.SetEnv()
		if err != nil {
			return err
		}
	}

	return nil
}
