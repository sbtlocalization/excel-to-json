// SPDX-FileCopyrightText: © 2026 SBT Localization https://sbt.localization.com.ua
// SPDX-FileContributor: Serhii Olendarenko <sergey.olendarenko@gmail.com>
//
// SPDX-License-Identifier: BlueOak-1.0.0

package main

import "runtime/debug"

// version is the Roman numeral of the release tag. The release build sets it
// with -ldflags "-X main.version=<tag>"; any other build leaves it empty.
var version string

// buildVersion is what --version prints. A build that did not come from a
// release tag never claims to be a release: it reports "dev", with the short
// commit it was built from when Go recorded one.
func buildVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" && len(setting.Value) >= 7 {
				return "dev+" + setting.Value[:7]
			}
		}
	}
	return "dev"
}
