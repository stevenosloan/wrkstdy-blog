package site

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

// bundleAssets builds the CSS and JS in root/assets with esbuild. Every .css
// file in assets/css and .js file in assets/js is an entry point; JS imports are
// resolved from root/node_modules. Output goes to out/assets under content-hashed
// names.
//
// The returned map is keyed like fingerprintDir's, "js/mermaid.js", and has an
// entry for each entry point, plus one for every font a stylesheet references
// ("fonts/fira-sans-latin-400-normal.woff2"), so templates can preload them.
// esbuild hashes the font, rewrites the url() to the hashed name, and hashes the
// CSS with that url() in it, so a changed font changes the CSS's name too.
//
// Code splitting is on, so a dynamic import() becomes a chunk that's fetched on
// demand, and its hash changes the importer's name too. That makes everything
// safe to cache forever. Only entry points and fonts are in the map, since the
// chunks are only ever loaded by the entry points.
func bundleAssets(root, out string) (map[string]Asset, error) {
	absRoot, absOut := mustAbs(root), mustAbs(out)
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
	if _, err := os.Stat(filepath.Join(root, "node_modules")); errors.Is(err, os.ErrNotExist) && hasJS(entries) {
		return nil, errors.New("no node_modules: run npm install (or npm ci) first")
	}

	outdir := filepath.Join(absOut, "assets")
	result := api.Build(api.BuildOptions{
		AbsWorkingDir: absRoot,
		EntryPoints:   entries,
		Outdir:        outdir,
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

		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		LegalComments:     api.LegalCommentsNone,
		Write:             false,
		LogLevel:          api.LogLevelSilent,
	})
	if len(result.Errors) > 0 {
		msgs := api.FormatMessages(result.Errors, api.FormatMessagesOptions{Kind: api.ErrorMessage})
		return nil, errors.New(strings.TrimSpace(strings.Join(msgs, "\n")))
	}

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
		m := hashedName.FindStringSubmatch(rel)
		if m == nil || strings.HasPrefix(rel, "js/chunks/") {
			continue
		}
		sum := sha256.Sum256(f.Contents)
		assets[m[1]+m[2]] = Asset{
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
