package onlineconf

import (
	"bytes"
	"encoding/json"
	"os"
	"path"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/colinmarc/cdb"
)

func getTmpFname(t *testing.T, pattern string) string {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		t.Fatalf(`os.CreateTemp("", %q): %v`, pattern, err)
	}

	f.Close()

	return f.Name()
}

func writeCDB(t *testing.T, fname string, tree map[string]string) {
	tmpFname := getTmpFname(t, "test_*.cdb.tmp")

	w, err := cdb.Create(tmpFname)
	if err != nil {
		t.Fatalf("cdb.Create(%s): %v", tmpFname, err)
	}

	childLists := map[string]map[string]struct{}{}

	for key, val := range tree {
		if err = w.Put(s2b(key), s2b("s"+val)); err != nil {
			t.Fatalf("cdb.Put(%q, %q): %v", key, val, err)
		}

		for dir := key; dir != "/"; {
			item := path.Base(dir)
			dir = path.Dir(dir)

			set, ok := childLists[dir]
			if !ok {
				childLists[dir] = map[string]struct{}{
					item: struct{}{},
				}
			} else {
				set[item] = struct{}{}
			}
		}
	}

	buf := &bytes.Buffer{}

	for dir, children := range childLists {
		buf.Reset()
		_, _ = buf.Write([]byte{'j'})

		list := make([]string, 0, len(children))
		for key := range children {
			list = append(list, key)
		}

		slices.Sort(list)

		if err = json.NewEncoder(buf).Encode(&list); err != nil {
			t.Fatalf("error encoding %s:%v: %v", dir, children, err)
		}

		buf.Truncate(buf.Len() - 1) // remove trailing linefeed

		listPath := "/"
		if dir != "/" {
			listPath = dir + "/"
		}

		if err = w.Put(s2b(listPath), buf.Bytes()); err != nil {
			t.Fatalf("cdb.Put(%q, %v): %v", listPath, children, err)
		}
	}

	if err = w.Close(); err != nil {
		t.Fatal("cdb.Close():", err)
	}

	if err = os.Rename(tmpFname, fname); err != nil {
		t.Fatalf("os.Rename(%q, %q): %v", tmpFname, fname, err)
	}
}

func waitChan(t *testing.T, key string, ch <-chan struct{}) {
	tm := time.NewTimer(time.Second)
	defer tm.Stop()

	select {
	case <-tm.C:
		t.Fatal(key, "subscription notification timed out")
	case <-ch:
	}
}

func getLongStr(n int) string {
	s := "qwerty"
	for range n {
		s += s
	}

	return s
}

func TestSubscriptions(t *testing.T) {
	initWatcherOnce = sync.OnceValues(initWatcherOnceFunc) // the watcher is getting lost during running the tests

	cdbName := getTmpFname(t, "test_*.cdb")
	defer os.Remove(cdbName)

	conf := map[string]string{
		"/test/key":         "val123",
		"/test/long":        getLongStr(9),
		"/test/subdir/key1": "val345",
		"/test/subdir/key2": "val678",
		"/test/unchanged":   "don't modify",
		"/test/deleted":     "", // test that empty and non-existent values are treated as different
	}

	writeCDB(t, cdbName, conf)

	mod, err := OpenModule(cdbName)
	if err != nil {
		t.Fatalf("OpenModule(%q): %v", cdbName, err)
	}

	shortCh, err := mod.Subscribe("/test/key")
	if err != nil {
		t.Fatal(`Subscribe("/test/key"):`, err)
	}

	shortCh2 := make(chan struct{}, 1)

	err = mod.SubscribeChan("/test/key", shortCh2)
	if err != nil {
		t.Fatal(`SubscribeChan("/test/key"):`, err)
	}

	longCh, err := mod.Subscribe("/test/long")
	if err != nil {
		t.Fatal(`Subscribe("/test/long"):`, err)
	}

	subdirCh, err := mod.SubscribeSubtree("/test/subdir")
	if err != nil {
		t.Fatal(`SubscribeSubdir("/test/subdir"):`, err)
	}

	unchangedCh, err := mod.Subscribe("/test/unchanged")
	if err != nil {
		t.Fatal(`Subscribe("/test/unchanged"):`, err)
	}

	deletedCh, err := mod.Subscribe("/test/deleted")
	if err != nil {
		t.Fatal(`Subscribe("/test/deleted"):`, err)
	}

	createdCh, err := mod.Subscribe("/test/created")
	if err != nil {
		t.Fatal(`Subscribe("/test/created"):`, err)
	}

	conf["/test/key"] = "changed"
	conf["/test/long"] = getLongStr(10)
	conf["/test/subdir/key1"] = "changed111"
	conf["/test/created"] = "hello"
	delete(conf, "/test/deleted")

	writeCDB(t, cdbName, conf)

	waitChan(t, "/test/key", shortCh)
	waitChan(t, "/test/key", shortCh2)
	waitChan(t, "/test/long", longCh)
	waitChan(t, "/test/subdir", subdirCh)
	waitChan(t, "/test/deleted", deletedCh)
	waitChan(t, "/test/created", createdCh)

	select {
	case <-unchangedCh:
		t.Fatal("/test/unchanged received a change notification")
	default:
	}
}
