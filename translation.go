// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"codeberg.org/tealeg/xlsx/v4"
)

// splitForms splits a translation at the gender separator. A translation that
// contains the separator always yields two forms, male then female, either of
// which may be empty; one without it yields the whole text as its only item.
func splitForms(value, separator string) ([]string, error) {
	if separator == "" {
		return nil, fmt.Errorf("gender separator must not be empty")
	}
	index := strings.Index(value, separator)
	if index < 0 {
		return []string{value}, nil
	}
	if strings.Contains(value[index+1:], separator) {
		return nil, fmt.Errorf("translation contains more than one gender separator")
	}
	return []string{value[:index], value[index+len(separator):]}, nil
}

// readTranslations returns the translation of every String ID in the first
// sheet of the spreadsheet, empty for an untranslated one. Columns are found by
// the names in the header row, in any order, and unrelated columns are ignored.
func readTranslations(path string) (map[int32]string, error) {
	if !strings.EqualFold(filepath.Ext(path), ".xlsx") {
		return nil, fmt.Errorf("translation file must be an XLSX file: %s", path)
	}
	wb, err := xlsx.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("open translation spreadsheet: %w", err)
	}
	if len(wb.Sheets) == 0 || wb.Sheets[0].MaxRow == 0 {
		return nil, fmt.Errorf("translation spreadsheet has no header")
	}
	sheet := wb.Sheets[0]
	header, err := sheet.Row(0)
	if err != nil {
		return nil, fmt.Errorf("read translation header: %w", err)
	}
	columns := map[string]int{}
	for i := 0; i < sheet.MaxCol; i++ {
		name := strings.ToLower(strings.TrimSpace(header.GetCell(i).Value))
		switch name {
		case "key", "source", "translation":
			if _, ok := columns[name]; ok {
				return nil, fmt.Errorf("duplicate column %q", name)
			}
			columns[name] = i
		}
	}
	for _, name := range []string{"key", "source", "translation"} {
		if _, ok := columns[name]; !ok {
			return nil, fmt.Errorf("missing required column %q", name)
		}
	}
	translations := map[int32]string{}
	for i := 1; i < sheet.MaxRow; i++ {
		row, err := sheet.Row(i)
		if err != nil {
			return nil, fmt.Errorf("read translation row %d: %w", i+1, err)
		}
		key := strings.TrimSpace(row.GetCell(columns["key"]).Value)
		source := row.GetCell(columns["source"]).Value
		translated := row.GetCell(columns["translation"]).Value
		if key == "" && source == "" && translated == "" {
			continue
		}
		id, err := strconv.ParseInt(key, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid String ID at row %d: %w", i+1, err)
		}
		if _, ok := translations[int32(id)]; ok {
			return nil, fmt.Errorf("duplicate String ID %d at row %d", id, i+1)
		}
		translations[int32(id)] = translated
	}
	return translations, nil
}
