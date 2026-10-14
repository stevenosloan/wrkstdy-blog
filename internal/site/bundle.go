package site

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

// Bundler builds the CSS and JS in root/assets with esbuild. Every .css file in
// assets/css and .js file in assets/js is an entry point; JS imports are resolved
// from root/node_modules. Output goes to out/assets.
//
// The map Build returns is keyed like fingerprintDir's, "js/mermaid.js", and has
// an entry for each entry point, plus one for every font a stylesheet references
// ("fonts/fira-sans-latin-400-normal.woff2"), so templates can preload them.
// Only those are in it, since chunks are only ever loaded by the entry points.
//
// A production Bundler names files by content hash, so they can be cached
// forever. esbuild hashes a font and rewrites the url() to the hashed name, then
// hashes the CSS with that url() in it, so a changed font changes the CSS's name
// too. Code splitting turns a dynamic import() into a chunk fetched on demand,
// and its hash changes the importer's name as well.
//
// A dev Bundler keeps an esbuild context between calls, so a rebuild only
// redoes what changed, and writes readable, unminified files with source maps
// and stable names for entry points and fonts, so devtools breakpoints survive a
// rebuild.
type Bundler struct {
	root, out string
	dev       bool
	entries   []string
	ctx       api.BuildContext // dev only: made on first use, and again if the entry points change
}

// NewBundler returns a Bundler for the project in root, writing into out.
func NewBundler(root, out string, dev bool) *Bundler {
	return &Bundler{root: root, out: out, dev: dev}
}

// Close frees the esbuild context a dev Bundler holds.
func (b *Bundler) Close() {
	if b.ctx != nil {
		b.ctx.Dispose()
		b.ctx = nil
	}
}

// Build runs esbuild and writes its output.
func (b *Bundler) Build() (map[string]Asset, error) {
	absRoot, absOut := mustAbs(b.root), mustAbs(b.out)
	assetsDir := filepath.Join(absRoot, "assets")

	var entries []string
	for _, pattern := range []string{"css/*.css", "js/*.js"} {
		matches, err := filepath.Glob(filepath.Join(assetsDir, filepath.FromSlash(pattern)))
		if err != nil {
			return nil, err
		}
		entries = append(entries, matches...)
	}
	if len(entries) == 0 {
		return map[string]Asset{}, nil
	}
	if _, err := os.Stat(filepath.Join(b.root, "node_modules")); errors.Is(err, os.ErrNotExist) && hasJS(entries) {
		return nil, errors.New("no node_modules: run npm install (or npm ci) first")
	}

	opts := api.BuildOptions{
		AbsWorkingDir: absRoot,
		EntryPoints:   entries,
		Outdir:        filepath.Join(absOut, "assets"),
		Outbase:       assetsDir,
		EntryNames:    "[dir]/[name].[hash]",
		ChunkNames:    "js/chunks/[name].[hash]",
		AssetNames:    "fonts/[name].[hash]",
		PublicPath:    "/assets",
		Loader:        map[string]api.Loader{".woff2": api.LoaderFile},
		Bundle:        true,
		Splitting:     true,
		Format:        api.FormatESModule,
		Target:        api.ES2022,
		Platform:      api.PlatformBrowser,
		LegalComments: api.LegalCommentsNone,
		Write:         false,
		LogLevel:      api.LogLevelSilent,
	}
	if b.dev {
		// Chunks keep their hash: esbuild names shared ones all "chunk", so they'd collide.
		// Only the entry points refer to them, and those are rewritten on every build.
		opts.EntryNames, opts.AssetNames = "[dir]/[name]", "fonts/[name]"
		opts.Sourcemap = api.SourceMapLinked
	} else {
		opts.MinifyWhitespace, opts.MinifyIdentifiers, opts.MinifySyntax = true, true, true
	}

	var result api.BuildResult
	if !b.dev {
		result = api.Build(opts)
	} else {
		if b.ctx != nil && !slices.Equal(entries, b.entries) {
			b.Close() // an entry point was added or removed
		}
		if b.ctx == nil {
			ctx, err := api.Context(opts)
			if err != nil {
				return nil, esbuildError(err.Errors)
			}
			b.ctx, b.entries = ctx, entries
		}
		result = b.ctx.Rebuild()
	}
	if len(result.Errors) > 0 {
		return nil, esbuildError(result.Errors)
	}

	outdir := opts.Outdir
	assets := map[string]Asset{}
	for _, f := range result.OutputFiles {
		if err := writeFile(f.Path, f.Contents); err != nil {
			return nil, err
		}
		rel, err := filepath.Rel(outdir, f.Path)
		if err != nil {
			return nil, err
		}
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, "js/chunks/") || strings.HasSuffix(rel, ".map") {
			continue
		}
		key := rel
		if !b.dev {
			m := hashedName.FindStringSubmatch(rel)
			if m == nil {
				continue
			}
			key = m[1] + m[2]
		}
		sum := sha256.Sum256(f.Contents)
		assets[key] = Asset{
			URL:       "/assets/" + rel,
			Integrity: "sha256-" + base64.StdEncoding.EncodeToString(sum[:]),
		}
	}
	for _, e := range entries {
		rel, _ := filepath.Rel(assetsDir, e)
		if _, ok := assets[filepath.ToSlash(rel)]; !ok {
			return nil, fmt.Errorf("esbuild wrote nothing for %s", rel)
		}
	}
	return assets, nil
}

func esbuildError(msgs []api.Message) error {
	text := api.FormatMessages(msgs, api.FormatMessagesOptions{Kind: api.ErrorMessage})
	return errors.New(strings.TrimSpace(strings.Join(text, "\n")))
}

// hashedName splits an output path like "js/mermaid.FYDPYMK6.js" into the name
// without the hash ("js/mermaid") and the extension (".js").
var hashedName = regexp.MustCompile(`^(.+)\.[A-Z0-9]{8}(\.[a-z0-9]+)$`)

func hasJS(entries []string) bool {
	for _, e := range entries {
		if filepath.Ext(e) == ".js" {
			return true
		}
	}
	return false
}

func mustAbs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return abs
}
