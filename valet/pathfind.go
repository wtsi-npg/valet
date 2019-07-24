/*
 * Copyright (C) 2019. Genome Research Ltd. All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License,
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * @file pathfind.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet

import (
	"context"
	"os"
	"path/filepath"
	"time"

	logs "github.com/kjsanger/logshim"
)

// FileResource is a locatable file.
type FileResource struct {
	Location string // Raw URL or file path
}

// FilePath is a FileResource that is on a local filesystem.
type FilePath struct {
	FileResource
	Info os.FileInfo
}

// NewFilePath returns a new instance where the path has been cleaned and made
// absolute and the FileInfo populated by os.Stat
func NewFilePath(path string) (FilePath, error) {
	var fp FilePath
	absPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return fp, err
	}

	info, err := os.Stat(absPath)
	fp.Info = info
	fp.FileResource = FileResource{absPath}

	return fp, err
}

// FindFiles walks the directory tree under root recursively, except into
// directories where pruneFn returns filepath.SkipDir, which prunes the
// directory traversal at that point.
//
// Files encountered are reported to the caller on the first returned (output)
// channel and any errors on the second (error) channel. Files are filtered by
// testing with the predicate pred; only where the predicate returns true are
// the files sent to the channel.
//
// The walking goroutine will continue to run until the directory tree is
// fully traversed, or the cancel function of cancelCtx is called. Either will
// close the output and error channels and exit the goroutine cleanly.
func FindFiles(
	ctx context.Context,
	root string,
	pred FilePredicate,
	pruneFn FilePredicate) (<-chan FilePath, <-chan error) {
	paths, errs := make(chan FilePath), make(chan error, 1)

	log := logs.GetLogger()
	log.Debug().Str("root", root).Msg("started find")

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				log.Warn().Err(err).Msg("file was deleted")
				return nil
			}

			return err
		}

		p := FilePath{FileResource{path}, info}

		if _, perr := pruneFn(p); perr != nil {
			if perr == filepath.SkipDir {
				log.Info().
					Str("path", path).
					Str("reason", perr.Error()).Msg("pruned path")
				return perr
			}
		}

		ok, perr := pred(p) // Predicate test

		if perr != nil {
			return perr
		} else if ok {
			log.Debug().Str("path", path).Msg("accepted")
			paths <- p
		} else {
			log.Debug().Str("path", path).Msg("rejected")
		}

		select {
		case <-ctx.Done():
			log.Debug().
				Str("root", root).
				Str("path", path).Msg("cancelled find")
			return nil
		default:
			return nil
		}
	}

	go func() {
		defer func() {
			close(paths)
			close(errs)
		}()

		root, rerr := filepath.Abs(root)
		if rerr != nil {
			errs <- rerr
		} else {
			werr := filepath.Walk(root, walkFn) // Directory walk
			if werr != nil {
				errs <- werr
			}
		}
	}()

	return paths, errs
}

// FindFilesInterval executes FindFiles every interval seconds. Aside from
// having the additional intervals parameter, it behaves in the same way as
// FindFiles.
func FindFilesInterval(
	ctx context.Context,
	root string, pred FilePredicate,
	pruneFn FilePredicate,
	interval time.Duration) (<-chan FilePath, <-chan error) {

	paths, errs := make(chan FilePath), make(chan error, 1)

	log := logs.GetLogger()
	findTick := time.NewTicker(interval)

	go func() {
		defer func() {
			close(paths)
			close(errs)
		}()

		for {
			select {
			case now := <-findTick.C:
				log.Debug().Str("root", root).
					Time("at", now).Msg("starting interval sweep")

				ipaths, ierrs := FindFiles(ctx, root, pred, pruneFn)
				for p := range ipaths {
					log.Debug().Msg("interval sweep sending path")
					paths <- p
				}
				for e := range ierrs {
					errs <- e
				}

			case <-ctx.Done():
				log.Debug().Str("root", root).
					Msg("cancelled interval sweep")
				findTick.Stop()
				return
			}
		}
	}()

	return paths, errs
}
