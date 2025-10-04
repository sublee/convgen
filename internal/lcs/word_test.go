package lcs_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sublee/convgen/internal/lcs"
)

func TestSplitWords(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "camelCase",
			input:    "getId",
			expected: []string{"get", "Id"},
		},
		{
			name:     "PascalCase",
			input:    "GetId",
			expected: []string{"Get", "Id"},
		},
		{
			name:     "snake_case",
			input:    "send_message",
			expected: []string{"send", "_", "message"},
		},
		{
			name:     "DigitAfterLetter",
			input:    "iso8601",
			expected: []string{"iso", "8601"},
		},
		{
			name:     "LetterAfterDigit",
			input:    "file2name",
			expected: []string{"file", "2", "name"},
		},
		{
			name:     "MultipleUnderscores",
			input:    "send__nowait",
			expected: []string{"send", "__", "nowait"},
		},
		{
			name:     "MixedCase",
			input:    "version2Point1",
			expected: []string{"version", "2", "Point", "1"},
		},
		{
			name:     "SingleWord",
			input:    "hello",
			expected: []string{"hello"},
		},
		{
			name:     "EmptyString",
			input:    "",
			expected: ([]string)(nil),
		},
		{
			name:     "AllUppercase",
			input:    "HELLO",
			expected: []string{"HELLO"},
		},
		{
			name:     "UppercaseAcronym",
			input:    "getID",
			expected: []string{"get", "ID"},
		},
		{
			name:     "UnderscoresOnly",
			input:    "___",
			expected: []string{"___"},
		},
		{
			name:     "DigitsOnly",
			input:    "12345",
			expected: []string{"12345"},
		},
		{
			name:     "Korean",
			input:    "안녕",
			expected: []string{"안녕"},
		},
		{
			name:     "UppercaseAcronymAtStart",
			input:    "JSONParser",
			expected: []string{"JSON", "Parser"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lcs.SplitWords(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestCommonWordPrefix(t *testing.T) {
	tt := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "camelCase",
			input:    []string{"getId", "getIdentifier", "getIndex"},
			expected: "get",
		},
		{
			name:     "camelCaseUpperAcronym",
			input:    []string{"getID", "getIdentifier", "getIndex"},
			expected: "get",
		},
		{
			name:     "camelCaseUpperAcronym2",
			input:    []string{"getID", "getIDd"},
			expected: "get",
		},
		{
			name:     "PascalCase",
			input:    []string{"GetID", "GetIdentifier", "GetIndex"},
			expected: "Get",
		},
		{
			name:     "snake_case",
			input:    []string{"send_nowait", "send_message", "send_data"},
			expected: "send_",
		},
		{
			name:     "DigitAfterLetter",
			input:    []string{"iso8601", "iso9000", "iso"},
			expected: "iso",
		},
		{
			name:     "LetterAfterDigit",
			input:    []string{"file2name", "file2path", "file2data"},
			expected: "file2",
		},
		{
			name:     "NoCommonPrefix",
			input:    []string{"firstName", "lastName", "middleName"},
			expected: "",
		},
		{
			name:     "EmptyInput",
			input:    []string{},
			expected: "",
		},
		{
			name:     "SingleString",
			input:    []string{"testFunction"},
			expected: "testFunction",
		},
		{
			name:     "IdenticalStrings",
			input:    []string{"testFunc", "testFunc", "testFunc"},
			expected: "testFunc",
		},
		{
			name:     "MixedCaseWithNumbers",
			input:    []string{"version2Point1", "version2Point2", "version2Point3"},
			expected: "version2Point",
		},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			got := lcs.CommonWordPrefix(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestCommonWordSuffix(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "camelCase",
			input:    []string{"getId", "fetchId", "removeId"},
			expected: "Id",
		},
		{
			name:     "PascalCase",
			input:    []string{"GetID", "FetchID", "RemoveID"},
			expected: "ID",
		},
		{
			name:     "snake_case",
			input:    []string{"send_message", "receive_message", "process_message"},
			expected: "_message",
		},
		{
			name:     "DigitAtEnd",
			input:    []string{"version1", "build1", "release1"},
			expected: "1",
		},
		{
			name:     "MixedCase",
			input:    []string{"handleErrorCode", "parseErrorCode", "formatErrorCode"},
			expected: "ErrorCode",
		},
		{
			name:     "NoCommonSuffix",
			input:    []string{"firstName", "lastName", "middleName"},
			expected: "Name",
		},
		{
			name:     "EmptyInput",
			input:    []string{},
			expected: "",
		},
		{
			name:     "SingleString",
			input:    []string{"testFunction"},
			expected: "testFunction",
		},
		{
			name:     "IdenticalStrings",
			input:    []string{"testFunc", "testFunc", "testFunc"},
			expected: "testFunc",
		},
		{
			name:     "ComplexSuffix",
			input:    []string{"oldVersion2Point1", "newVersion2Point1", "nextVersion2Point1"},
			expected: "Version2Point1",
		},
		{
			name:     "PartialWordMatch",
			input:    []string{"getData", "setData", "Data"},
			expected: "Data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lcs.CommonWordSuffix(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
