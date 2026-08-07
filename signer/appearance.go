package signer

import (
	"fmt"
	"os"
	"path/filepath"

	pdfsign "github.com/digitorus/pdfsign"
)

// Appearance describes an optional visual signature widget placed on a page.
//
// When ImagePath is empty, the widget uses pdfsign's standard text layout
// (signer name, reason, location, and date). When ImagePath is set, the
// image fills the widget instead.
//
// pdfsign only allows visible appearances on Options.Type ==
// pdfsign.ApprovalSignature; SignFile returns an error otherwise.
type Appearance struct {
	// Page is the 1-indexed page the widget is placed on. Defaults to 1.
	Page int `mapstructure:"page"`
	// X, Y is the lower-left corner of the widget, in Unit units.
	X float64 `mapstructure:"x"`
	Y float64 `mapstructure:"y"`
	// Width, Height is the size of the widget, in Unit units.
	Width  float64 `mapstructure:"width"`
	Height float64 `mapstructure:"height"`
	// Unit scales X, Y, Width, and Height (e.g. pdfsign.Millimeter). Defaults
	// to PDF points (1.0) when zero.
	Unit float64 `mapstructure:"unit"`
	// ImagePath, if set, is a PNG/JPEG file drawn to fill the widget instead
	// of the standard text layout.
	ImagePath string `mapstructure:"image"`
}

// build renders a into a pdfsign.Appearance ready to hand to SignBuilder.Appearance.
func (a *Appearance) build() (*pdfsign.Appearance, error) {
	app := pdfsign.NewAppearance(a.Width, a.Height)

	if a.ImagePath == "" {
		app.Standard()
		return app, nil
	}

	data, err := os.ReadFile(a.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("read appearance image: %w", err)
	}

	app.Image(&pdfsign.Image{Name: filepath.Base(a.ImagePath), Data: data}).
		Rect(0, 0, a.Width, a.Height).
		ScaleFit()

	return app, nil
}

// page returns the 1-indexed page the widget is placed on, defaulting to 1.
func (a *Appearance) page() int {
	if a.Page <= 0 {
		return 1
	}

	return a.Page
}

// unit returns the coordinate scale, defaulting to PDF points (1.0).
func (a *Appearance) unit() float64 {
	if a.Unit <= 0 {
		return 1
	}

	return a.Unit
}
