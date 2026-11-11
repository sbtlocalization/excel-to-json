// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"codeberg.org/tealeg/xlsx/v4"
)

func execute(args ...string) (string, string, error) {
	cmd := newRootCommand()
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return stdout.String(), stderr.String(), err
}

func translationSheet(t *testing.T, rows [][]string) string {
	t.Helper()
	wb := xlsx.NewFile()
	sheet, err := wb.AddSheet("Translation")
	if err != nil {
		t.Fatal(err)
	}
	for _, values := range rows {
		row := sheet.AddRow()
		for _, value := range values {
			row.AddCell().SetString(value)
		}
	}
	path := filepath.Join(t.TempDir(), "translation.xlsx")
	if err := wb.Save(path); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestConvertProducesTextJSON(t *testing.T) {
	sheet := translationSheet(t, [][]string{
		{"translation", "extra", " Source ", " KEY "},
		{"translated", "ignored", "original", "1"},
		{"", "ignored", "source /// untouched", "2"},
		{"male /// female", "", "", "3"},
		{"", "", "", "4"},
	})
	stdout, stderr, err := execute("-i", sheet)
	if err != nil {
		t.Fatalf("convert: %v; stderr: %s", err, stderr)
	}
	var got map[string][]string
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("stdout is not JSON: %q: %v", stdout, err)
	}
	want := map[string][]string{"1": {"translated"}, "3": {"male", "female"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("JSON = %#v, want %#v", got, want)
	}
	if stderr != "" {
		t.Errorf("stderr = %q", stderr)
	}
}

func TestConvertGenderFormsAndWhitespace(t *testing.T) {
	for _, tc := range []struct {
		name, separator string
		rows            [][]string
		want            map[string][]string
	}{
		{
			name: "default separator",
			rows: [][]string{
				{"1", "", " /// female"}, {"2", "", "male /// "},
				{"3", "", " /// "}, {"4", "", "  shared  "},
				{"5", "", " male  ///  female "}, {"6", "", "same /// same"},
				{"7", "", "Командир /// "},
			},
			want: map[string][]string{
				"1": {"", "female"}, "2": {"male", ""}, "3": {"", ""}, "4": {"  shared  "},
				"5": {" male ", " female "}, "6": {"same", "same"}, "7": {"Командир", ""},
			},
		},
		{
			name: "custom separator", separator: "<>",
			rows: [][]string{{"1", "", "  male <> female  "}, {"2", " source<>not split ", ""}, {"3", "", "<>"}, {"4", "", "male <> "}},
			want: map[string][]string{"1": {"  male ", " female  "}, "3": {"", ""}, "4": {"male ", " "}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rows := append([][]string{{"key", "source", "translation"}}, tc.rows...)
			sheet := translationSheet(t, rows)
			args := []string{"-i", sheet}
			if tc.separator != "" {
				args = append(args, "--separator", tc.separator)
			}
			stdout, stderr, err := execute(args...)
			if err != nil {
				t.Fatalf("convert: %v; stderr: %s", err, stderr)
			}
			var got map[string][]string
			if err := json.Unmarshal([]byte(stdout), &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("JSON = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestConvertEmptyAndFirstSheet(t *testing.T) {
	wb := xlsx.NewFile()
	first, err := wb.AddSheet("First")
	if err != nil {
		t.Fatal(err)
	}
	for _, values := range [][]string{{"key", "source", "translation", "extra"}, {"", "", "", "ignored"}, {"42", "untranslated source", "", "ignored"}} {
		row := first.AddRow()
		for _, value := range values {
			row.AddCell().SetString(value)
		}
	}
	second, err := wb.AddSheet("Second")
	if err != nil {
		t.Fatal(err)
	}
	second.AddRow().AddCell().SetString("invalid header")
	sheet := filepath.Join(t.TempDir(), "two.xlsx")
	if err := wb.Save(sheet); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, err := execute("-i", sheet)
	if err != nil || strings.TrimSpace(stdout) != "{}" || stderr != "" {
		t.Fatalf("stdout = %q, stderr = %q, err = %v; want empty object", stdout, stderr, err)
	}
}

func TestConvertRejectsInvalidSpreadsheets(t *testing.T) {
	for _, tc := range []struct {
		name string
		rows [][]string
		want string
	}{
		{"missing key", [][]string{{"source", "translation"}}, "key"},
		{"missing source", [][]string{{"key", "translation"}}, "source"},
		{"missing translation", [][]string{{"key", "source"}}, "translation"},
		{"duplicate header", [][]string{{"key", "source", "translation", " KEY "}}, "duplicate column"},
		{"invalid ID", [][]string{{"key", "source", "translation"}, {"abc", "source", ""}}, "row 2"},
		{"missing ID", [][]string{{"key", "source", "translation"}, {"", "source", ""}}, "row 2"},
		{"overflow", [][]string{{"key", "source", "translation"}, {"2147483648", "", "text"}}, "row 2"},
		{"underflow", [][]string{{"key", "source", "translation"}, {"-2147483649", "", "text"}}, "row 2"},
		{"duplicate untranslated", [][]string{{"key", "source", "translation"}, {"5", "one", ""}, {"5", "two", ""}}, "duplicate String ID 5"},
		{"duplicate empty", [][]string{{"key", "source", "translation"}, {"5", "", ""}, {"05", "", "translation"}}, "duplicate String ID 5"},
		{"duplicate translated", [][]string{{"key", "source", "translation"}, {"5", "", "one"}, {"5", "", "two"}}, "duplicate String ID 5"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sheet := translationSheet(t, tc.rows)
			stdout, stderr, err := execute("-i", sheet)
			if err == nil || !strings.Contains(stderr, tc.want) || stdout != "" {
				t.Fatalf("stdout = %q, stderr = %q, err = %v; want error containing %q", stdout, stderr, err, tc.want)
			}
		})
	}
}

func TestConvertRejectsInvalidInputFiles(t *testing.T) {
	dir := t.TempDir()
	invalid := filepath.Join(dir, "invalid.xlsx")
	if err := os.WriteFile(invalid, []byte("not a spreadsheet"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{invalid, filepath.Join(dir, "missing.xlsx"), filepath.Join(dir, "wrong.txt")} {
		stdout, stderr, err := execute("-i", path)
		if err == nil || stderr == "" || stdout != "" {
			t.Errorf("input %s: stdout = %q, stderr = %q, err = %v", path, stdout, stderr, err)
		}
	}
	stdout, stderr, err := execute()
	if err == nil || !strings.Contains(stderr, "input") || stdout != "" {
		t.Errorf("missing input: stdout = %q, stderr = %q, err = %v", stdout, stderr, err)
	}
}

func TestConvertValidationDoesNotWriteOutput(t *testing.T) {
	sheet := translationSheet(t, [][]string{{"key", "source", "translation"}, {"1", "", "valid"}, {"2", "", "one /// two /// three"}})
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"multiple separators", nil, "String ID 2"},
		{"empty separator", []string{"--separator", ""}, "separator must not be empty"},
		{"overlapping custom separators", []string{"--separator", "aa"}, "String ID 2"},
		{"invalid ID", nil, "row 3"},
		{"duplicate ID", nil, "duplicate String ID 1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := sheet
			switch tc.name {
			case "overlapping custom separators":
				input = translationSheet(t, [][]string{{"key", "source", "translation"}, {"2", "", "aaa"}})
			case "invalid ID":
				input = translationSheet(t, [][]string{{"key", "source", "translation"}, {"1", "", "valid"}, {"bad", "", "text"}})
			case "duplicate ID":
				input = translationSheet(t, [][]string{{"key", "source", "translation"}, {"1", "", "valid"}, {"1", "source", ""}})
			}
			stdout, stderr, err := execute(append([]string{"-i", input}, tc.args...)...)
			if err == nil || !strings.Contains(stderr, tc.want) || stdout != "" {
				t.Fatalf("stdout = %q, stderr = %q, err = %v", stdout, stderr, err)
			}
			for _, existing := range []bool{false, true} {
				path := filepath.Join(t.TempDir(), "nested", "out.json")
				if existing {
					if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(path, []byte("keep me"), 0644); err != nil {
						t.Fatal(err)
					}
				}
				stdout, stderr, err := execute(append([]string{"-i", input, "-o", path}, tc.args...)...)
				if err == nil || !strings.Contains(stderr, tc.want) || stdout != "" {
					t.Fatalf("stdout = %q, stderr = %q, err = %v", stdout, stderr, err)
				}
				data, readErr := os.ReadFile(path)
				if existing && (readErr != nil || string(data) != "keep me") {
					t.Fatalf("existing output = %q, err = %v", data, readErr)
				}
				if !existing && !os.IsNotExist(readErr) {
					t.Fatalf("output unexpectedly exists: %v", readErr)
				}
			}
		})
	}
}

func TestConvertWritesNestedOutput(t *testing.T) {
	sheet := translationSheet(t, [][]string{{"key", "source", "translation"}, {"-2147483648", "source", "minimum"}, {"2147483647", "", "translated"}})
	path := filepath.Join(t.TempDir(), "nested", "deeper", "result.json")
	stdout, stderr, err := execute("--input", sheet, "--output", path)
	if err != nil || stdout != "" || stderr != "" {
		t.Fatalf("stdout = %q, stderr = %q, err = %v", stdout, stderr, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string][]string
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	want := map[string][]string{"-2147483648": {"minimum"}, "2147483647": {"translated"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("JSON = %#v, want %#v", got, want)
	}
}

func TestHasNoSubcommands(t *testing.T) {
	if got := newRootCommand().Commands(); len(got) != 0 {
		t.Errorf("root command has subcommands: %v", got)
	}
}

func TestVersionPrintsTheTag(t *testing.T) {
	previous := version
	t.Cleanup(func() { version = previous })
	version = "IV"
	stdout, stderr, err := execute("--version")
	if err != nil || stdout != "IV\n" || stderr != "" {
		t.Fatalf("stdout = %q, stderr = %q, err = %v; want \"IV\\n\"", stdout, stderr, err)
	}
}

func TestVersionOfAnUntaggedBuildIsNotARelease(t *testing.T) {
	previous := version
	t.Cleanup(func() { version = previous })
	version = ""
	got := buildVersion()
	if got != "dev" && !strings.HasPrefix(got, "dev+") {
		t.Errorf("buildVersion() = %q, want dev or dev+<commit>", got)
	}
}
