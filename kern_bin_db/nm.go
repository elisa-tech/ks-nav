package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type symtype int

const (
	Unknown symtype = iota
	TEXT
	DATA
	BSS
	COMMON
	SMALL_INITIALIZED
	ABSOLUTE
	RO
	WEAK_SYMBOL
	WEAK_OBJECT
)

func nmTypeToEnum(t rune) symtype {
	switch t {
	case 'A', 'a':
		return ABSOLUTE
	case 'B', 'b':
		return BSS
	case 'C', 'c':
		return COMMON
	case 'D', 'd':
		return DATA
	case 'G', 'g':
		return SMALL_INITIALIZED
	case 'R', 'r':
		return RO
	case 'T', 't':
		return TEXT
	case 'W', 'w':
		return WEAK_SYMBOL
	case 'V', 'v':
		return WEAK_OBJECT
	default:
		return Unknown
	}
}

type Symbol struct {
	Address uint64
	Type    symtype
	Name    string
}

func ParseSymbol(line string) (*Symbol, error) {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid line format")
	}

	address, err := strconv.ParseUint(parts[0], 16, 64)
	if err != nil {
		return nil, err
	}

	symType := nmTypeToEnum(rune(parts[1][0]))
	if symType == Unknown {
		return nil, fmt.Errorf("Unknown symbol type");
	}

	return &Symbol{
		Address: address,
		Type:    symType,
		Name:    parts[2],
	}, nil
}

func GetSymbols(toolchainPrefix string, file string) ([]Symbol, error) {
	cmd := exec.Command(toolchainPrefix+"nm", "-n", file)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(&out)
	var symbols []Symbol

	for scanner.Scan() {
		symbol, err := ParseSymbol(scanner.Text())
		if err != nil {
			continue
		}
		symbols = append(symbols, *symbol)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return symbols, nil
}
