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
	logs "github.com/kjsanger/logshim"
	"github.com/pkg/errors"

	"github.com/kjsanger/valet/utilities"
)

// WatchFiles reports filesystem events on the directories below root. Watches
// are set up recursively on every directory, except those for which pruneFn
// returns filepath.SkipDir, which prunes the directory traversal at that
// point. WatchFiles uses an internal event handler to add watches to any new
// directories added to the tree while is is operating, except those pruned as
// described.
//
// Events on files are reported to the caller on the first returned (output)
// channel and any errors on the second (error) channel. Events are filtered by
// testing the event file with the predicate pred; only where the predicate
// returns true are the events sent to the channel.
//
// The watching goroutine will continue to run until the cancel function of
// cancelCtx is called. This will close the output and error channels and exit
// the goroutine cleanly.
func WatchFiles(
	cancelCtx context.Context,
	root string,
	pred FilePredicate,
	pruneFn FilePredicate) (<-chan FilePath, <-chan error) {

	paths, errs := make(chan FilePath), make(chan error, 1)

	watcher, err := setupWatcher(root, pruneFn)
	if err != nil {
		errs <- err
		return paths, errs
	}

	log := logs.GetLogger()
	log.Info().Str("root", root).Msg("started watch")

	watchFn := func(ctx context.Context) (ferr error) {
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
					if os.IsNotExist(ferr) {
						log.Warn().Err(ferr).Msg("path deleted externally")
						return nil
					}

					return ferr
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					if ferr = handleCreateDir(p, pruneFn, watcher); ferr != nil {
						return ferr
					}
				}
				if event.Op&fsnotify.Close == fsnotify.Close {
					if ferr = handleCloseFile(p, pred, paths); ferr != nil {
						return ferr
					}
				}
				if event.Op&fsnotify.Movedto == fsnotify.Movedto {
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

func setupWatcher(root string,
	prune FilePredicate) (watcher *fsnotify.Watcher, err error) {
	if err := ensureIsDir(root); err != nil {
		return nil, err
	}

	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		if watcher != nil {
			return watcher, utilities.CombineErrors(err, watcher.Close())
		}

		return watcher, err
	}

	log := logs.GetLogger()

	// Pruning/skipping in go works by throwing the special error SkipDir, This
	// means that the main return value of the prune predicate is ignored.
	// Therefore we can say
	//
	// include := And(IsDir, prune)
	//
	// or
	//
	// include := And(IsDir, Not(prune))
	//
	include := And(IsDir, Not(prune))

	walkFn := func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		fp := FilePath{FileResource{path}, info}
		ok, err := include(fp)
		if err != nil {
			return err
		}

		if ok {
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

func handleCreateDir(target FilePath, prune FilePredicate,
	watcher *fsnotify.Watcher) error {

	log := logs.GetLogger()
	log.Debug().
		Str("path", target.Location).
		Str("op", "Create").Msg("handled event")

	ok, err := And(IsDir, Not(prune))(target)
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

	log := logs.GetLogger()
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

	log := logs.GetLogger()
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
