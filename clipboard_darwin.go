// +build darwin

package clipboard

import (
	"errors"

	"git.wow.st/gmp/clip"
)

func set(text string) error {
	ok := clip.Set(text)
	if !ok {
		return errors.New("nothing to set string")
	}
	return nil
}

func get() (string, error) {
	return clip.Get(), nil
}
