// Package signer signs and verifies PDF documents on behalf of pdfsigner,
// built on pdfsign's fluent Document API.
package signer

import (
	"crypto"
	"fmt"
	"os"

	pdfsign "github.com/digitorus/pdfsign"
	"github.com/digitorus/pdfsigner/license"
	log "github.com/sirupsen/logrus"
)

// Options configures how a signature is produced: metadata, format,
// timestamping, and an optional visual appearance.
type Options struct {
	Reason     string `mapstructure:"reason"`
	Location   string `mapstructure:"location"`
	Contact    string `mapstructure:"contactInfo"`
	SignerName string `mapstructure:"name"`

	Type       pdfsign.SignatureType `mapstructure:"certType"`
	Permission pdfsign.Permission    `mapstructure:"docMDP"`
	Format     pdfsign.Format        `mapstructure:"format"`
	// Digest selects the hash algorithm. Zero keeps pdfsign's default (SHA256).
	Digest crypto.Hash `mapstructure:"-"`

	TSAURL      string `mapstructure:"tsaUrl"`
	TSAUsername string `mapstructure:"tsaUsername"`
	TSAPassword string `mapstructure:"tsaPassword"`

	// Appearance, when set, draws a visual signature widget on the page.
	Appearance *Appearance `mapstructure:"appearance"`
}

// SignFile signs input with identity according to opts, writing the result
// to output. It blocks on the license rate limiter before signing, and
// optionally re-verifies the result when validate is true.
func SignFile(input, output string, identity *Identity, opts Options, validate bool) error {
	if err := license.LD.Wait(); err != nil {
		return err
	}

	if err := signFile(input, output, identity, opts); err != nil {
		return err
	}

	if validate {
		if _, err := VerifyFile(output); err != nil {
			return fmt.Errorf("validate signed output: %w", err)
		}
	}

	log.Println("File signed:", output)

	return nil
}

func signFile(input, output string, identity *Identity, opts Options) error {
	doc, err := pdfsign.OpenFile(input)
	if err != nil {
		return fmt.Errorf("open %s: %w", input, err)
	}

	sb := doc.Sign(identity.Signer, identity.Certificate, identity.Intermediates...).
		Reason(opts.Reason).
		Location(opts.Location).
		Contact(opts.Contact).
		SignerName(opts.SignerName).
		Type(opts.Type).
		Permission(opts.Permission).
		Format(opts.Format)

	if opts.Digest != 0 {
		sb.Digest(opts.Digest)
	}

	if opts.TSAURL != "" {
		sb.Timestamp(opts.TSAURL).TimestampAuth(opts.TSAUsername, opts.TSAPassword)
	}

	if opts.Appearance != nil {
		app, err := opts.Appearance.build()
		if err != nil {
			return err
		}

		sb.Unit(opts.Appearance.unit()).
			Appearance(app, opts.Appearance.X, opts.Appearance.Y).
			Page(opts.Appearance.page())
	}

	out, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("create %s: %w", output, err)
	}
	defer func() { _ = out.Close() }()

	if _, err := doc.Write(out); err != nil {
		return fmt.Errorf("sign %s: %w", input, err)
	}

	return nil
}
