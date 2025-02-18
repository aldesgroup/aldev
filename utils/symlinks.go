// ----------------------------------------------------------------------------
// Creates the required symlinks, if missing
// ----------------------------------------------------------------------------
package utils

import (
	"os"
	"path/filepath"
	"time"
)

// Checks if the required symlinks exist, else creates them
func CreateMissingSymlinks(ctx CancelableContext) {
	if len(Config().Symlinks) > 0 {
		start := time.Now()
		for _, symlinkCfg := range Config().Symlinks {
			createMissingSymlink(ctx, symlinkCfg)
		}
		Debug("Checked / created the missing symlinks in '%s'", time.Since(start))
	} else {
		Debug("No symlink to create")
	}
}

func createMissingSymlink(ctx CancelableContext, symlinkCfg *SymlinkConfig) {
	// checking the symlink existence
	info, err := os.Lstat(symlinkCfg.As)
	if err != nil {
		if os.IsNotExist(err) {
			Debug("Symlink '%s' does not exist.", symlinkCfg.As)
		} else {
			FatalIfErr(ctx, err)
		}
	}

	// if it does exist, exiting here
	if info != nil && info.Mode()&os.ModeSymlink != 0 {
		Debug("Symlink '%s' already exists", symlinkCfg.As)
		return
	}

	// using absolute paths to avoid errors
	link, errLink := filepath.Abs(symlinkCfg.Link)
	FatalIfErr(ctx, errLink)
	as, errAs := filepath.Abs(symlinkCfg.As)
	FatalIfErr(ctx, errAs)

	// creating the missing directories if needed
	Debug("Making sure '%s' exists (for '%s')", filepath.Dir(as), as)
	EnsureDir(ctx, filepath.Dir(as))

	// creation
	Debug("Creating symlink: '%s' --> '%s'", as, link)
	FatalIfErr(ctx, os.Symlink(link, as))
}
