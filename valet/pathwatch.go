/*
 * Copyright (C) 2019, 2020, 2021, 2022 Genome Research Ltd. All rights
 * reserved.
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

	"github.com/pkg/errors"
	"github.com/wtsi-npg/fsnotify"
	logs "github.com/wtsi-npg/logshim"

	"github.com/wtsi-npg/valet/utilities"
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

	// Buffer any error that may occur starting the watcher, so that
	// we can send it to the channel without blocking WatchFiles from returning
	paths, errs := make(chan FilePath), make(chan error, 1)
	log := logs.GetLogger()

	// Returns an error on failure to finish cleanly. Errors encountered
	// while running are sent to the error channel.
	watchFn := func(ctx context.Context, w *fsnotify.Watcher) (ferr error) { // NRV
		log.Info().Str("root", root).Msg("started watch")
		defer func() {
			if err := w.Close(); err != nil {
				ferr = utilities.CombineErrors(ferr, err)
			}
		}()

		for {
			select {
			case event := <-w.Events:
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					// Don't try to create FilePaths for removed files

					// Note: in the currently used version of fsnotify,
					// duplicate events are generated for removal. You will see
					// this message twice
					log.Info().Str("path", event.Name).
						Msg("was removed")
					continue
				}

				var p FilePath
				p, err := NewFilePath(event.Name)
				if err != nil {
					if os.IsNotExist(err) {
						log.Warn().Err(err).Str("path", event.Name).
							Msg("path deleted externally")
						continue
					}

					errs <- errors.WithMessagef(err,
						"on event: %s", event.Name)
					continue
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					if err = handleCreateDir(p, pruneFn, w); err != nil {
						errs <- errors.WithMessagef(err,
							"on directory creation: %s", p.Location)
					}
				}
				if event.Op&fsnotify.Close == fsnotify.Close {
					if ferr = handleCloseFile(p, pred, paths); ferr != nil {
						errs <- errors.WithMessagef(err,
							"on file close: %s", p.Location)
					}
				}
				if event.Op&fsnotify.Movedto == fsnotify.Movedto {
					if ferr = handleMovedtoFile(p, pred, paths); ferr != nil {
						errs <- errors.WithMessagef(err,
							"on file move: %s", p.Location)
					}
				}

			case <-ctx.Done():
				log.Info().Str("root", root).Msg("cancelled watch")
				return

			case err := <-w.Errors:
				errs <- err
			}
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		watcher = nil
		errs <- err
	}

	go func() {
		defer func() {
			close(paths)
			close(errs)
		}()

		if watcher != nil {
			// We might encounter an error before adding all possible dirs. It's
			// better for production if we handle the error by logging and press
			// on. The logs will be monitored. Meanwhile, data may still load
			// via the filesystem sweeps.
			if err := addWatchDirs(watcher, root, pruneFn); err != nil {
				errs <- err
			}
			if err := watchFn(cancelCtx, watcher); err != nil {
				errs <- err
			}
		}
	}()

	return paths, errs
}

func addWatchDirs(watcher *fsnotify.Watcher, root string, prune FilePredicate) error {
	if err := ensureIsDir(root); err != nil {
		return err
	}

	log := logs.GetLogger()

	// Pruning/skipping in go works by throwing the special error SkipDir, This
	// means that the main return value of the prune predicate is ignored.
	// Therefore, we can say
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
			log.Error().Err(walkErr).Str("path", path).
				Msg("when setting up initial watches")
			return nil
		}

		fp := FilePath{FileResource{path}, info}
		ok, err := include(fp)
		if err != nil {
			return err
		}

		if ok {
			walkErr = watcher.Add(path)
			if walkErr == nil {
				log.Info().Str("path", path).Msg("added to watcher")
			}
		}

		return walkErr
	}

	return filepath.Walk(root, walkFn)
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
				Msg("added to watcher")
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
			Msg("accepted by WatchFiles")
		paths <- target
	} else {
		log.Debug().
			Str("path", target.Location).
			Msg("rejected by WatchFiles")
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
			Msg("accepted by WatchFiles")
		paths <- target
	} else {
		log.Debug().
			Str("path", target.Location).
			Msg("rejected by WatchFiles")
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
