// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package main

import "os"

func main() {
	if err := newRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
