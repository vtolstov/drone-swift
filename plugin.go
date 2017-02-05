package main

import (
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/h2non/filetype"
	"github.com/mattn/go-zglob"
	"github.com/ncw/swift"
)

// Plugin defines the swift plugin parameters.
type Plugin struct {
	Endpoint    string
	Key         string
	Secret      string
	Container   string
	AuthVersion int
	Region      string
	Tenant      string
	Timeout     string
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

	//Exclude path
	Exclude []string

	// Strip the prefix from the target path
	StripPrefix string

	// Dry run without uploading/
	DryRun bool

	conn *swift.Connection
}

// Exec runs the plugin
func (p *Plugin) Exec() error {

	// create the client
	conn := &swift.Connection{
		UserName:    p.Key,
		ApiKey:      p.Secret,
		AuthUrl:     p.Endpoint,
		AuthVersion: p.AuthVersion,
	}

	if p.AuthVersion > 1 {
		conn.Region = p.Region
		conn.Tenant = p.Tenant
	}

	if td, err := time.ParseDuration(p.Timeout); err == nil {
		conn.Timeout = td
	}

	if err := conn.Authenticate(); err != nil {
		return err
	}

	p.conn = conn
	logrus.WithFields(logrus.Fields{
		"region":    p.Region,
		"endpoint":  p.Endpoint,
		"container": p.Container,
		"path":      p.Target,
	}).Info("Attempting to upload")

	sources, err := matches(p.Source, p.Exclude)
	if err != nil {
		return err
	}

	for _, source := range sources {
		if err = filepath.Walk(source, p.walk()); err != nil {
			return err
		}
	}
	return nil
}

func (p *Plugin) walk() filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		p.uploadFile(path)
		return nil
	}
}

// uploadFile is the helper function to upload file
func (p *Plugin) uploadFile(source string) error {
	content := contentType(source)

	target := filepath.Join(p.Target, strings.TrimPrefix(source, p.StripPrefix))
	// log file for debug purposes.
	logrus.WithFields(logrus.Fields{
		"source":       source,
		"container":    p.Container,
		"target":       target,
		"content-type": content,
	}).Info("Uploading file")

	// when executing a dry-run we exit because we don't actually want to
	// upload the file to swift.
	if p.DryRun {
		return nil
	}

	fr, err := os.Open(source)
	if err != nil {
		return err
	}
	defer fr.Close()

	fw, err := p.conn.ObjectCreate(p.Container, target, false, "", content, nil)
	if err != nil {
		return err
	}
	if _, err = io.Copy(fw, fr); err != nil {
		return err
	}
	if err = fw.Close(); err != nil {
		return err
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
		if len(matches) < 1 {
			return nil, os.ErrNotExist
		}
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

	if len(included) < 1 {
		return nil, os.ErrNotExist
	}

	return included, nil
}

// contentType is a helper function that returns the content type for the file
// based on extension. If the file extension is unknown application/octet-stream
// is returned.
func contentType(path string) string {
	ftype := "application/octet-stream"
	r, err := os.Open(path)
	if err != nil {
		return ftype
	}
	defer r.Close()
	dtype, err := filetype.MatchReader(r)
	if err != nil {
		return ftype
	}
	if dtype.MIME.Type != "" && dtype.MIME.Value != "" {
		return dtype.MIME.Value
	}
	ext := filepath.Ext(path)
	if ftype := mime.TypeByExtension(ext); ftype != "" {
		return ftype
	}
	return "application/octet-stream"
}
