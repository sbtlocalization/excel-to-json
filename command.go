// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "excel-to-json -i <XLSX file>",
		Short: "Convert a translation spreadsheet to Text JSON",
		Long: `Turns a translation spreadsheet into Text JSON for use with other tools.
Only translated rows are taken, and a translation may hold both gender forms.`,
		Args:         cobra.NoArgs,
		RunE:         runConvert,
		SilenceUsage: true,
		Version:      buildVersion(),
	}
	cmd.SetVersionTemplate("{{.Version}}\n")
	cmd.Flags().StringP("input", "i", "", "translation spreadsheet `file`")
	cmd.Flags().StringP("output", "o", "", "output JSON `file` (default: standard output)")
	cmd.Flags().String("separator", " /// ", "gender `separator` between male and female forms")
	cmd.MarkFlagRequired("input")
	cmd.MarkFlagFilename("input", "xlsx")
	cmd.MarkFlagFilename("output", "json")
	return cmd
}

func runConvert(cmd *cobra.Command, _ []string) error {
	input, _ := cmd.Flags().GetString("input")
	output, _ := cmd.Flags().GetString("output")
	separator, _ := cmd.Flags().GetString("separator")
	if separator == "" {
		return fmt.Errorf("gender separator must not be empty")
	}
	translations, err := readTranslations(input)
	if err != nil {
		return err
	}
	result := make(map[string][]string)
	for id, text := range translations {
		if text == "" {
			continue
		}
		forms, err := splitForms(text, separator)
		if err != nil {
			return fmt.Errorf("String ID %d: %w", id, err)
		}
		result[strconv.FormatInt(int64(id), 10)] = forms
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("encode Text JSON: %w", err)
	}
	data = append(data, '\n')
	if output == "" {
		_, err := cmd.OutOrStdout().Write(data)
		return err
	}
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(output, data, 0644); err != nil {
		return fmt.Errorf("write Text JSON: %w", err)
	}
	return nil
}
