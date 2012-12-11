/*
Copyright 2011-2012 Paul Ruane.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"tmsu/common"
	"tmsu/database"
	"tmsu/fingerprint"
)

func ValidateTagName(tagName string) error {
	if tagName == "." || tagName == ".." {
		return errors.New("Tag name cannot be '.' or '..'.")
	}

	if strings.Index(tagName, ",") != -1 {
		return errors.New("Tag names cannot contain commas.")
	}

	if strings.Index(tagName, "=") != -1 {
		return errors.New("Tag names cannot contain '='.")
	}

	if strings.Index(tagName, " ") != -1 {
		return errors.New("Tag names cannot contain spaces.")
	}

	if strings.Index(tagName, "/") != -1 {
		return errors.New("Tag names cannot contain slashes.")
	}

	if tagName[0] == '-' {
		return errors.New("Tag names cannot start with '-'.")
	}

	return nil
}

func AddFile(db *database.Database, path string) (*database.File, error) {
	info, err := os.Stat(path)
	if err != nil {
		switch {
		case os.IsPermission(err):
			return nil, errors.New(fmt.Sprintf("'%v': Permisison denied", path))
		case os.IsNotExist(err):
			return nil, errors.New(fmt.Sprintf("'%v': No such file", path))
		default:
			return nil, errors.New(fmt.Sprintf("'%v': Error: %v", path, err))
		}
	}
	modTime := info.ModTime().UTC()

	fingerprint, err := fingerprint.Create(path)
	if err != nil {
		return nil, err
	}

	file, err := db.FileByPath(path)
	if err != nil {
		return nil, err
	}

	if file == nil {
		// new file

		if !info.IsDir() {
			duplicateCount, err := db.FileCountByFingerprint(fingerprint)
			if err != nil {
				return nil, err
			}

			if duplicateCount > 0 {
				common.Warn("'" + common.RelPath(path) + "' is a duplicate file.")
			}
		}

		file, err = db.AddFile(path, fingerprint, modTime)
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			fsFile, err := os.Open(file.Path())
			if err != nil {
				return nil, err
			}
			defer fsFile.Close()

			dirFilenames, err := fsFile.Readdirnames(0)
			if err != nil {
				return nil, err
			}

			for _, dirFilename := range dirFilenames {
				_, err = AddFile(db, filepath.Join(path, dirFilename))
				if err != nil {
					return nil, err
				}
			}
		}
	} else {
		// existing file

		if file.ModTimestamp.Unix() != modTime.Unix() {
			db.UpdateFile(file.Id, file.Path(), fingerprint, modTime)
		}
	}

	return file, nil
}