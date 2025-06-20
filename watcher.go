package onlineconf

import (
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/fsnotify/fsnotify"
)

const tracebackMaxSize = 65536

var initWatcherOnce = sync.OnceValues(initWatcherOnceFunc)

func initWatcherOnceFunc() (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("fsnotify.NewWatcher: %w", err)
	}

	go func() {
		for {
			func() {
				defer func() {
					if reason := recover(); reason != nil {
						log.Printf("watcher panic: %v\n%s\n", reason, traceback())
					}
				}()

				select {
				case ev := <-watcher.Events:
					// log.Println("fsnotify event:", ev)
					if ev.Op&fsnotify.Create == fsnotify.Create {
						module, ok := modCache.loadOnly(ev.Name) // paths are always absolute
						if !ok {
							break
						}

						if err := module.reopen(); err != nil {
							log.Printf("watch %s: reopen failed: %v", ev.Name, err)
						}
					}

				case err := <-watcher.Errors:
					log.Print("Watch error: ", err)
				}
			}()
		}
	}()

	return watcher, nil
}

func initWatcher(dir string) error {
	watcher, err := initWatcherOnce()
	if err != nil {
		return err
	}

	if err = watcher.Add(dir); err != nil { // watcher.Add is goro-safe
		return fmt.Errorf("fsnotify.Watcher.Add: %w", err)
	}

	return nil
}

func traceback() string {
	traceback := make([]byte, tracebackMaxSize)
	size := runtime.Stack(traceback, false)

	return b2s(traceback[:size])
}
