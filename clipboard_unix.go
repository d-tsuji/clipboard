// The MIT License (MIT)
// Copyright (c) 2016 Alessandro Arzilli
// https://github.com/aarzilli/nucular/blob/master/LICENSE

// +build freebsd linux netbsd openbsd solaris dragonfly

package clipboard

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/sevlyar/go-daemon"
)

const debugClipboardRequests = false

var (
	x             *xgb.Conn
	win           xproto.Window
	clipboardText string
	selnotify     chan bool

	clipboardAtom, primaryAtom, textAtom, targetsAtom, atomAtom xproto.Atom
	targetAtoms                                                 []xproto.Atom
	clipboardAtomCache                                          = map[xproto.Atom]string{}

	doneCh = make(chan interface{})
)

func start() error {
	var err error
	xServer := os.Getenv("DISPLAY")
	if xServer == "" {
		return errors.New("could not identify xserver")
	}
	x, err = xgb.NewConnDisplay(xServer)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	selnotify = make(chan bool, 1)

	win, err = xproto.NewWindowId(x)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	setup := xproto.Setup(x)
	s := setup.DefaultScreen(x)
	err = xproto.CreateWindowChecked(x, s.RootDepth, win, s.Root, 100, 100, 1, 1, 0, xproto.WindowClassInputOutput, s.RootVisual, 0, []uint32{}).Check()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	clipboardAtom = internAtom(x, "CLIPBOARD")
	primaryAtom = internAtom(x, "PRIMARY")
	textAtom = internAtom(x, "UTF8_STRING")
	targetsAtom = internAtom(x, "TARGETS")
	atomAtom = internAtom(x, "ATOM")

	targetAtoms = []xproto.Atom{targetsAtom, textAtom}

	go eventLoop()

	return nil
}

func set(text string) error {
	d := &daemon.Context{}

	_, err := d.Reborn()
	if err != nil {
		fmt.Errorf("unable to run: %w", err)
	}
	defer d.Release()

	//　The following will run as a daemon.
	if err := start(); err != nil {
		return fmt.Errorf("init clipboard: %w", err)
	}
	clipboardText = text
	ssoc := xproto.SetSelectionOwnerChecked(x, win, clipboardAtom, xproto.TimeCurrentTime)
	if err := ssoc.Check(); err != nil {
		return fmt.Errorf("setting clipboard: %w", err)
	}

	// Wait for the SelectionClear event.
	<-doneCh

	return nil
}

func get() (string, error) {
	if err := start(); err != nil {
		return "", fmt.Errorf("init clipboard: %w", err)
	}
	return getSelection(clipboardAtom)
}

func getSelection(selAtom xproto.Atom) (string, error) {
	csc := xproto.ConvertSelectionChecked(x, win, selAtom, textAtom, selAtom, xproto.TimeCurrentTime)
	err := csc.Check()
	if err != nil {
		return "", fmt.Errorf("convert selection check: %w", err)
	}

	select {
	case r := <-selnotify:
		if !r {
			return "", nil
		}
		gpc := xproto.GetProperty(x, true, win, selAtom, textAtom, 0, 5*1024*1024)
		gpr, err := gpc.Reply()
		if err != nil {
			return "", fmt.Errorf("grp reply: %w", err)
		}
		if gpr.BytesAfter != 0 {
			return "", errors.New("clipboard too large")
		}
		return string(gpr.Value[:gpr.ValueLen]), nil
	case <-time.After(1 * time.Second):
		return "", errors.New("clipboard retrieval failed, timeout")
	}
}

func eventLoop() {
	for {
		e, err := x.WaitForEvent()
		if err != nil {
			fmt.Fprintln(os.Stderr, "WaitForEvent error")
			continue
		}

		switch e := e.(type) {
		case xproto.SelectionRequestEvent:
			if debugClipboardRequests {
				tgtname := lookupAtom(e.Target)
				fmt.Fprintln(os.Stderr, "SelectionRequest", e, textAtom, tgtname, "isPrimary:", e.Selection == primaryAtom, "isClipboard:", e.Selection == clipboardAtom)
			}
			t := clipboardText

			switch e.Target {
			case textAtom:
				if debugClipboardRequests {
					fmt.Fprintln(os.Stderr, "Sending as text")
				}
				cpc := xproto.ChangePropertyChecked(x, xproto.PropModeReplace, e.Requestor, e.Property, textAtom, 8, uint32(len(t)), []byte(t))
				if cpc.Check() == nil {
					sendSelectionNotify(e)
				} else {
					fmt.Fprintln(os.Stderr, err)
				}

			case targetsAtom:
				if debugClipboardRequests {
					fmt.Fprintln(os.Stderr, "Sending targets")
				}
				buf := make([]byte, len(targetAtoms)*4)
				for i, atom := range targetAtoms {
					xgb.Put32(buf[i*4:], uint32(atom))
				}

				cpc := xproto.ChangePropertyChecked(x, xproto.PropModeReplace, e.Requestor, e.Property, atomAtom, 32, uint32(len(targetAtoms)), buf)
				if cpc.Check() == nil {
					sendSelectionNotify(e)
				} else {
					fmt.Fprintln(os.Stderr, err)
				}

			default:
				if debugClipboardRequests {
					fmt.Fprintln(os.Stderr, "Skipping")
				}
				e.Property = 0
				sendSelectionNotify(e)
			}
		case xproto.SelectionNotifyEvent:
			selnotify <- (e.Property == clipboardAtom) || (e.Property == primaryAtom)
		case xproto.SelectionClearEvent:
			// Client loses ownership of a selection, so daemon process exit.
			doneCh <- struct{}{}
		}
	}
}

func lookupAtom(at xproto.Atom) string {
	if s, ok := clipboardAtomCache[at]; ok {
		return s
	}

	reply, err := xproto.GetAtomName(x, at).Reply()
	if err != nil {
		panic(err)
	}

	// If we're here, it means we didn't have ths ATOM id cached. So cache it.
	atomName := string(reply.Name)
	clipboardAtomCache[at] = atomName
	return atomName
}

func sendSelectionNotify(e xproto.SelectionRequestEvent) {
	sn := xproto.SelectionNotifyEvent{
		Time:      e.Time,
		Requestor: e.Requestor,
		Selection: e.Selection,
		Target:    e.Target,
		Property:  e.Property}
	sec := xproto.SendEventChecked(x, false, e.Requestor, 0, string(sn.Bytes()))
	err := sec.Check()
	if err != nil {
		fmt.Println(err)
	}
}

func internAtom(conn *xgb.Conn, n string) xproto.Atom {
	iac := xproto.InternAtom(conn, true, uint16(len(n)), n)
	iar, err := iac.Reply()
	if err != nil {
		panic(err)
	}
	return iar.Atom
}
