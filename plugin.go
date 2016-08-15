package main

import (
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/mattn/go-zglob"
	"github.com/ncw/swift"
)

// Plugin defines the S3 plugin parameters.
type Plugin struct {
	Endpoint    string
	Key         string
	Secret      string
	Container   string
	AuthVersion int
	Region      string
	Tenant      string
	// Copies the files from the specified directory.
	// Regexp matching will apply to match multiple
	// files
	//
	// Examples:
	//    /path/to/file
	//    /path/to/*.txt
	//    /path/to/*/*.txt
	//    /path/to/**
	Source string
	Target string

	// Strip the prefix from the target path
	StripPrefix string

	// Recursive uploads
	Recursive bool

	// Exclude files matching this pattern.
	Exclude []string

	// Dry run without uploading/
	DryRun bool
}

// Exec runs the plugin
func (p *Plugin) Exec() error {

	// create the client
	conn := &swift.Connection{
		UserName: p.Key,
		ApiKey:   p.Secret,
		AuthUrl:  p.Endpoint,
	}

	if p.AuthVersion > 1 {
		conn.Region = p.Region
		conn.Tenant = p.Tenant
	}

	err := conn.Authenticate()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Could not match files")
		return err
	}

	log.WithFields(log.Fields{
		"region":    p.Region,
		"endpoint":  p.Endpoint,
		"container": p.Container,
	}).Info("Attempting to upload")

	matches, err := matches(p.Source, p.Exclude)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Could not match files")
		return err
	}

	for _, match := range matches {

		stat, err := os.Stat(match)
		if err != nil {
			continue // should never happen
		}

		// skip directories
		if stat.IsDir() {
			continue
		}

		target := filepath.Join(p.Target, strings.TrimPrefix(match, p.StripPrefix))
		if !strings.HasPrefix(target, "/") {
			target = "/" + target
		}

		content := contentType(match)

		// log file for debug purposes.
		log.WithFields(log.Fields{
			"name":         match,
			"container":    p.Container,
			"target":       target,
			"content-type": content,
		}).Info("Uploading file")

		// when executing a dry-run we exit because we don't actually want to
		// upload the file to swift.
		if p.DryRun {
			continue
		}

		fr, err := os.Open(match)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"file":  match,
			}).Error("Problem opening file")
			return err
		}
		defer fr.Close()

		fw, err := conn.ObjectCreate(p.Container, target, false, "", content, nil)
		if err != nil {
			log.WithFields(log.Fields{
				"name":      match,
				"container": p.Container,
				"target":    target,
				"error":     err,
			}).Error("Could not upload file")
			return err
		}
		if _, err = io.Copy(fw, fr); err != nil {
			log.WithFields(log.Fields{
				"name":      match,
				"container": p.Container,
				"target":    target,
				"error":     err,
			}).Error("Could not upload file")
			return err
		}
		if err = fw.Close(); err != nil {
			log.WithFields(log.Fields{
				"name":      match,
				"container": p.Container,
				"target":    target,
				"error":     err,
			}).Error("Could not upload file")
			return err
		}

		fr.Close()
	}

	return nil
}

// matches is a helper function that returns a list of all files matching the
// included Glob pattern, while excluding all files that match the exclusion
// Glob pattners.
func matches(include string, exclude []string) ([]string, error) {
	matches, err := zglob.Glob(include)
	if err != nil {
		return nil, err
	}
	if len(exclude) == 0 {
		return matches, nil
	}

	// find all files that are excluded and load into a map. we can verify
	// each file in the list is not a member of the exclusion list.
	excludem := map[string]bool{}
	for _, pattern := range exclude {
		excludes, err := zglob.Glob(pattern)
		if err != nil {
			return nil, err
		}
		for _, match := range excludes {
			excludem[match] = true
		}
	}

	var included []string
	for _, include := range matches {
		_, ok := excludem[include]
		if ok {
			continue
		}
		included = append(included, include)
	}
	return included, nil
}

// contentType is a helper function that returns the content type for the file
// based on extension. If the file extension is unknown application/octet-stream
// is returned.
func contentType(path string) string {
	ext := filepath.Ext(path)
	typ := mime.TypeByExtension(ext)
	if typ == "" {
		typ = "application/octet-stream"
	}
	return typ
}
