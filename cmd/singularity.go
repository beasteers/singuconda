package cmd

import (
	"compress/gzip"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
)

var DEFAULT_SING_NAME = GetEnvVar("SING_CMD", "sing")

func GetOverlay() (string, string, string, error) {
	singName := DEFAULT_SING_NAME

	// look for existing overlays in this directory
	existingMatches, err := filepath.Glob("*.ext3")
	if err != nil {
		return "", "", singName, err
	}

	// select from existing overlays
	if len(existingMatches) > 0 {
		prompt1 := promptui.Select{
			Label: "There are overlays in this directory. Use one?",
			Items: append(existingMatches, "new..."),
		}
		_, existingOverlay, err := prompt1.Run()
		if err != nil {
			return "", "", singName, err
		}
		if existingOverlay != "new..." {
			overlayName := strings.TrimSuffix(path.Base(existingOverlay), ".gz")
			overlayName = strings.TrimSuffix(path.Base(overlayName), filepath.Ext(overlayName))
			// existingOverlay, _ = filepath.Abs(existingOverlay)
			return existingOverlay, overlayName, singName, nil
		}
	}

	// select new overlay
	matches, err := filepath.Glob(filepath.Join(OVERLAY_DIR, "*.ext3.gz"))
	if err != nil {
		return "", "", singName, err
	}

	searcher := func(input string, index int) bool {
		name := strings.Replace(strings.ToLower(matches[index]), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(name, input)
	}

	prompt2 := promptui.Select{
		Label:             "Which overlay to use?",
		Items:             matches,
		Searcher:          searcher,
		StartInSearchMode: true,
		CursorPos:         indexOf(filepath.Join(OVERLAY_DIR, DEFAULT_OVERLAY), matches),
	}
	_, overlayPath, err := prompt2.Run()
	if err != nil {
		return "", "", singName, err
	}

	// give the overlay a new name
	defaultOverlayName := path.Base(overlayPath)
	defaultOverlayName = strings.TrimSuffix(defaultOverlayName, ".gz")
	defaultOverlayName = strings.TrimSuffix(defaultOverlayName, filepath.Ext(defaultOverlayName))
	prompt3 := promptui.Prompt{
		Label:   "Why don't you give your overlay a name?",
		Default: defaultOverlayName,
	}
	name, err := prompt3.Run()
	if err != nil {
		return "", "", singName, err
	}
	if name == "" {
		name = defaultOverlayName
	}
	fmt.Printf("You choose %q\n", name)

	overlayDest := fmt.Sprintf("%s.ext3", name)
	if _, err := os.Stat(overlayDest); !os.IsNotExist(err) {
		fmt.Printf("file exists %s\n", overlayDest)
		return "", "", singName, err
	}

	// expand the overlay to the current directory
	fmt.Printf("Unzipping %s to %s...\n", overlayPath, overlayDest)
	f, err := os.Open(overlayPath)
	if err != nil {
		return "", "", singName, err
	}
	reader, err := gzip.NewReader(f)
	if err != nil {
		return "", "", singName, err
	}
	defer reader.Close()

	o, err := os.Create(overlayDest)
	if err != nil {
		return "", "", singName, err
	}
	defer o.Close()
	_, err = o.ReadFrom(reader)
	if err != nil {
		return "", "", singName, err
	}
	fmt.Printf("Done!\n")
	return overlayDest, name, singName, nil
}

func GetSif(name string) (string, error) {
	sifCache := fmt.Sprintf(".%s.sifpath", name)

	// check if we configured the sif file before
	defaultSif := filepath.Join(SIF_DIR, DEFAULT_SIF)
	if _, err := os.Stat(sifCache); err == nil {
		buf, err := os.ReadFile(sifCache)
		if err != nil {
			return "", err
		}
		defaultSif = string(buf)

		promptyn := promptui.Prompt{
			Label:     fmt.Sprintf("Use %s", defaultSif),
			IsConfirm: true,
			Default:   "y",
		}
		_, err = promptyn.Run()
		if err == nil {
			return defaultSif, nil
		}
	}

	// select from sifs
	matches, err := filepath.Glob(filepath.Join(SIF_DIR, "*.sif"))
	if err != nil {
		return "", err
	}

	searcher := func(input string, index int) bool {
		name := strings.Replace(strings.ToLower(matches[index]), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)
		return strings.Contains(name, input)
	}
	prompt := promptui.Select{
		Label:             "Which sif to use?",
		Items:             matches,
		Searcher:          searcher,
		StartInSearchMode: true,
		CursorPos:         indexOf(defaultSif, matches),
	}
	_, sifPath, err := prompt.Run()
	if err != nil {
		return "", err
	}

	// write cache for next time
	err = os.WriteFile(sifCache, []byte(sifPath), 0774)
	if err != nil {
		return "", err
	}

	return sifPath, nil
}
