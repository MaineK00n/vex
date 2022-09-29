/*
Copyright 2022 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package ctl

import (
	"fmt"

	"chainguard.dev/vex/pkg/attestation"
	"chainguard.dev/vex/pkg/sarif"
	"chainguard.dev/vex/pkg/vex"
)

type VexCtl struct {
	impl    Implementation
	Options Options
}

type Options struct {
	Products []string // List of products to match in CSAF docs
	Format   string   // Firmat of the vex documents
	Sign     bool     // When true, attestations will be signed before attaching
}

func New() *VexCtl {
	return &VexCtl{
		impl: &defaultVexCtlImplementation{},
	}
}

// ApplyFiles takes a list of paths to vex files and applies them to a report
func (vexctl *VexCtl) ApplyFiles(r *sarif.Report, files []string) (*sarif.Report, error) {
	vexes, err := vexctl.impl.OpenVexData(vexctl.Options, files)
	if err != nil {
		return nil, fmt.Errorf("opening vex data: %w", err)
	}

	return vexctl.Apply(r, vexes)
}

// Apply takes a sarif report and applies one or more vex documents
func (vexctl *VexCtl) Apply(r *sarif.Report, vexDocs []*vex.VEX) (finalReport *sarif.Report, err error) {
	// Sort the docs by date
	vexDocs = vexctl.impl.Sort(vexDocs)

	// Apply the sorted documents to the report
	for i, doc := range vexDocs {
		finalReport, err = vexctl.impl.ApplySingleVEX(r, doc)
		if err != nil {
			return nil, fmt.Errorf("applying vex document #%d: %w", i, err)
		}
	}

	return finalReport, nil
}

// Generate an attestation from a VEX
func (vexctl *VexCtl) Attest(vexDataPath string, imageRefs []string) (*attestation.Attestation, error) {
	doc, err := vexctl.impl.OpenVexData(vexctl.Options, []string{vexDataPath})
	if err != nil {
		return nil, fmt.Errorf("opening vex data")
	}

	// Generate the attestation
	att := attestation.New()
	att.Predicate = doc
	att.AddImageSubjects(imageRefs)

	return att, nil
}

// Attach attaches an attestation to a list of images
func (vexctl *VexCtl) Attach(att *attestation.Attestation, imageRefs []string) (err error) {
	var attestationBytes []byte
	if vexctl.Options.Sign {
		attestationBytes, err = vexctl.impl.SignAttestation(att)
	} else {
		attestationBytes, err = vexctl.impl.AttestationBytes(att)
	}
	if err != nil {
		return fmt.Errorf("signing attestation: %w", err)
	}

	fmt.Println(string(attestationBytes))
	return nil
}
