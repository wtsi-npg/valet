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
 * @file pathwatch.go
 * @author Keith James <kdj@sanger.ac.uk>
 */

package valet

import (
	"context"
	"os"
	"path/filepath"

	"github.com/kjsanger/fsnotify"
	"github.com/pkg/errors"
	logf "valet/log/logfacade"
	"valet/utilities"
)

func WatchFiles(
	cancelCtx context.Context,
	root string,
	pred FilePredicate) (<-chan FilePath, <-chan error) {

	paths, errs := make(chan FilePath), make(chan error, 1)

	watcher, err := setupWatcher(root)
	if err != nil {
		errs <- err
		return paths, errs
	}

	log := logf.GetLogger()
	log.Info().Str("root", root).Msg("started watch")

	watchFn := func(ctx context.Context) error {
		var ferr error
		defer func() {
			if werr := watcher.Close(); werr != nil {
				ferr = utilities.CombineErrors(ferr, werr)
			}
		}()

		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					// Don't try to create FilePaths for removed files
					continue
				}

				p, ferr := NewFilePath(event.Name)
				if ferr != nil {
					return ferr
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					if ferr = handleCreateDir(p, watcher); ferr != nil {
						return ferr
					}
				} else if event.Op&fsnotify.Close == fsnotify.Close {
					if ferr = handleCloseFile(p, pred, paths); ferr != nil {
						return ferr
					}
				} else if event.Op&fsnotify.Movedto == fsnotify.Movedto {
					if ferr = handleMovedtoFile(p, pred, paths); ferr != nil {
						return ferr
					}
				}

			case <-ctx.Done():
				log.Info().Str("root", root).Msg("cancelled watch")
				return nil

			case ferr = <-watcher.Errors:
				return ferr
			}
		}
	}

	go func() {
		defer func() {
			close(paths)
			close(errs)
		}()

		err = watchFn(cancelCtx)
		if err != nil {
			errs <- err
		}
	}()

	return paths, errs
}

func setupWatcher(root string) (*fsnotify.Watcher, error) {
	if err := ensureIsDir(root); err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		if watcher != nil {
			return watcher, utilities.CombineErrors(err, watcher.Close())
		}

		return watcher, err
	}

	log := logf.GetLogger()

	walkFn := func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if info.IsDir() {
			walkErr = watcher.Add(path)
			if walkErr == nil {
				log.Info().Str("path", path).Msg("added watcher")
			}
		}

		return walkErr
	}

	err = filepath.Walk(root, walkFn)

	return watcher, err
}

func handleCreateDir(target FilePath, watcher *fsnotify.Watcher) error {

	log := logf.GetLogger()
	log.Debug().
		Str("path", target.Location).
		Str("op", "Create").Msg("handled event")

	ok, err := IsDir(target)
	if err != nil {
		return err
	}
	if ok {
		err = watcher.Add(target.Location)
		if err == nil {
			log.Info().
				Str("path", target.Location).
				Msg("added watcher")
		}
	}
	return err
}

func handleCloseFile(target FilePath, pred FilePredicate,
	paths chan FilePath) error {

	log := logf.GetLogger()
	log.Debug().
		Str("path", target.Location).
		Str("op", "Close").Msg("handled event")

	ok, err := pred(target)
	if err != nil {
		return err
	}
	if ok {
		log.Debug().
			Str("path", target.Location).
			Msg("accepted for processing")
		paths <- target
	} else {
		log.Debug().
			Str("path", target.Location).
			Msg("rejected (predicate false)")
	}

	return err
}

func handleMovedtoFile(target FilePath, pred FilePredicate,
	paths chan FilePath) error {

	log := logf.GetLogger()
	log.Debug().
		Str("path", target.Location).
		Str("op", "Movedto").Msg("handled event")

	ok, err := pred(target)
	if err != nil {
		return err
	}
	if ok {
		log.Debug().
			Str("path", target.Location).
			Msg("accepted for processing")
		paths <- target
	} else {
		log.Debug().
			Str("path", target.Location).
			Msg("rejected (predicate false)")
	}

	return err
}

func ensureIsDir(path string) error {
	fInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !fInfo.IsDir() {
		return errors.Errorf("%s was not a directory", path)
	}
	return err
}
