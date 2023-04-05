/*
 * Copyright (c) 2022 Red Hat, Inc.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */
package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nav Tests", func() {
	Describe("opt2num", func() {
		When("Using a valid options key", func() {
			It("Should return the correct value for graphOnly", func() {
				Expect(opt2num("graphOnly")).To(Equal(graphOnly))
			})

			It("Should return the correct value for jsonOutputPlain", func() {
				Expect(opt2num("jsonOutputPlain")).To(Equal(jsonOutputPlain))
			})

			It("Should return the correct value for jsonOutputB64", func() {
				Expect(opt2num("jsonOutputB64")).To(Equal(jsonOutputB64))
			})

			It("Should return the correct value for jsonOutputGZB64", func() {
				Expect(opt2num("jsonOutputGZB64")).To(Equal(jsonOutputGZB64))
			})
		})

		When("Using an invalid options key", func() {
			It("Should return 0", func() {
				Expect(opt2num("invalidKey")).To(Equal(invalidOutput))
			})
		})
	})

	Describe("decorateLine", func() {
		var l string
		var r string
		var adjm []adjM
		BeforeEach(func() {
			l = "lsys"
			r = "rsys"
			adjm = []adjM{
				{
					l: node{
						subsys:     "lsys",
						symbol:     "lsym",
						sourceRef:  "lsource",
						addressRef: "laddr",
					},
					r: node{
						subsys:     "rsys",
						symbol:     "rsym",
						sourceRef:  "rsource",
						addressRef: "raddr",
					},
				},
			}
		})
		When("Subsystems match", func() {
			It("Should return a list of all matched subsystems", func() {
				actual := decorateLine(l, r, adjm)
				expected := " [label=\"rsym([raddr]rsource),\\n\"]"

				Expect(actual).To(Equal(expected))
			})

			It("Should ignore duplicated entries", func() {
				duplicated := adjM{
					l: node{
						subsys:     "lsys",
						symbol:     "lsym",
						sourceRef:  "lsource",
						addressRef: "laddr",
					},
					r: node{
						subsys:     "rsys",
						symbol:     "rsym",
						sourceRef:  "rsource",
						addressRef: "raddr",
					},
				}
				adjm = append(adjm, duplicated)

				actual := decorateLine(l, r, adjm)
				expected := " [label=\"rsym([raddr]rsource),\\n\"]"

				Expect(actual).To(Equal(expected))
			})

			It("Should return more than one entry", func() {
				extra := adjM{
					l: node{
						subsys:     "lsys",
						symbol:     "lsym2",
						sourceRef:  "lsource2",
						addressRef: "laddr2",
					},
					r: node{
						subsys:     "rsys",
						symbol:     "rsym2",
						sourceRef:  "rsource2",
						addressRef: "raddr2",
					},
				}
				adjm = append(adjm, extra)

				actual := decorateLine(l, r, adjm)
				expected := " [label=\"rsym([raddr]rsource),\\nrsym2([raddr2]rsource2),\\n\"]"

				Expect(actual).To(Equal(expected))
			})
		})

		When("Subsystems do not match", func() {
			It("Should return an empty list if using an empty slice", func() {
				actual := decorateLine(l, r, []adjM{})
				expected := " [label=\"\"]"

				Expect(actual).To(Equal(expected))

			})

			It("Should return an empty list if nodes do not match", func() {
				actual := decorateLine(l, "asym", adjm)
				expected := " [label=\"\"]"

				Expect(actual).To(Equal(expected))

			})
		})
	})

	Describe("generateOutput", func() {
		// TODO: `nav.generateOutput` refactoring needed
	})

	Describe("main", func() {
		// TODO: `nav.main` refactoring needed
	})
})
