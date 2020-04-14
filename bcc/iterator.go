package bcc

// Iterator iterates over some type of value. It's used much like a
// bufio.Scanner or an sql.Rows.
//
//    for iter.Next() {
//      cur := iter.Current()
//
//      // ...
//    }
//    if err := iter.Err(); err != nil {
//      // ...
//    }
type Iterator struct {
	next  func() bool
	cur   func() (interface{}, error)
	close func() error

	cache interface{}
	err   error
}

// Next advances the iterator to the next value. The iterator starts
// before the first value, so this much be called once before anything
// else. It returns false if there is no next value to advance to.
func (iter *Iterator) Next() bool {
	if iter.err != nil {
		return false
	}

	more := iter.next()
	if !more {
		return false
	}

	v, err := iter.cur()
	if err != nil {
		iter.cache = nil
		iter.err = err
		return false
	}
	iter.cache = v

	return true
}

func (iter *Iterator) Close() error {
	return iter.close()
}

// Current returns the current value of the iteration. This value is
// cached during the call to Next, so this is a cheap call.
func (iter *Iterator) Current() interface{} {
	return iter.cache
}

func (iter *Iterator) Err() error {
	return iter.err
}
