package bolt

import (
	"errors"
	"github.com/ansible-semaphore/semaphore/db"
)

func (d *BoltDb) SetOption(key string, value string) error {

	opt := db.Option{
		Key:   key,
		Value: value,
	}

	_, err := d.getOption(key)

	if errors.Is(err, db.ErrNotFound) {
		_, err = d.createObject(-1, db.OptionProps, opt)
		return err
	} else {
		err = d.updateObject(-1, db.OptionProps, opt)
	}

	return err
}

func (d *BoltDb) getOption(key string) (value string, err error) {
	var option db.Option
	err = d.getObject(-1, db.OptionProps, strObjectID(key), &option)
	value = option.Value
	return
}

func (d *BoltDb) GetOption(key string) (value string, err error) {
	var option db.Option
	err = d.getObject(-1, db.OptionProps, strObjectID(key), &option)
	value = option.Value

	if errors.Is(err, db.ErrNotFound) {
		err = nil
	}

	return
}
