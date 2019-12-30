package diff

import (
	"bytes"
	"fmt"
	"io"
	"reflect"

	"github.com/pkg/diff/myers"
	"github.com/pkg/diff/write"
)

// DiffWrite is the union of myers.Pair and write.Pair:
// It can be diffed using myers diff, and written in unified diff format.
type DiffWrite interface {
	myers.Pair
	write.Pair
}

// Strings returns a DiffWrite that can diff and write a and b.
func Strings(a, b []string) DiffWrite {
	return &diffStrings{a: a, b: b}
}

type diffStrings struct {
	a, b []string
}

func (ab *diffStrings) LenA() int                                { return len(ab.a) }
func (ab *diffStrings) LenB() int                                { return len(ab.b) }
func (ab *diffStrings) Equal(ai, bi int) bool                    { return ab.a[ai] == ab.b[bi] }
func (ab *diffStrings) WriteATo(w io.Writer, i int) (int, error) { return io.WriteString(w, ab.a[i]) }
func (ab *diffStrings) WriteBTo(w io.Writer, i int) (int, error) { return io.WriteString(w, ab.b[i]) }

// Bytes returns a DiffWrite that can diff and write a and b.
func Bytes(a, b [][]byte) DiffWrite {
	return &diffBytes{a: a, b: b}
}

type diffBytes struct {
	a, b [][]byte
}

func (ab *diffBytes) LenA() int                                { return len(ab.a) }
func (ab *diffBytes) LenB() int                                { return len(ab.b) }
func (ab *diffBytes) Equal(ai, bi int) bool                    { return bytes.Equal(ab.a[ai], ab.b[bi]) }
func (ab *diffBytes) WriteATo(w io.Writer, i int) (int, error) { return w.Write(ab.a[i]) }
func (ab *diffBytes) WriteBTo(w io.Writer, i int) (int, error) { return w.Write(ab.b[i]) }

// Slices returns a DiffWrite that diffs a and b.
// It uses fmt.Print to print the elements of a and b.
// It uses equal to compare elements of a and b;
// if equal is nil, Slices uses reflect.DeepEqual.
func Slices(a, b interface{}, equal func(x, y interface{}) bool) DiffWrite {
	if equal == nil {
		equal = reflect.DeepEqual
	}
	ab := &diffSlices{a: reflect.ValueOf(a), b: reflect.ValueOf(b), eq: equal}
	if ab.a.Type().Kind() != reflect.Slice || ab.b.Type().Kind() != reflect.Slice {
		panic(fmt.Errorf("diff.Slices called with a non-slice argument: %T, %T", a, b))
	}
	return ab
}

type diffSlices struct {
	a, b reflect.Value
	eq   func(x, y interface{}) bool
}

func (ab *diffSlices) LenA() int                                { return ab.a.Len() }
func (ab *diffSlices) LenB() int                                { return ab.b.Len() }
func (ab *diffSlices) atA(i int) interface{}                    { return ab.a.Index(i).Interface() }
func (ab *diffSlices) atB(i int) interface{}                    { return ab.b.Index(i).Interface() }
func (ab *diffSlices) Equal(ai, bi int) bool                    { return ab.eq(ab.atA(ai), ab.atB(bi)) }
func (ab *diffSlices) WriteATo(w io.Writer, i int) (int, error) { return fmt.Fprint(w, ab.atA(i)) }
func (ab *diffSlices) WriteBTo(w io.Writer, i int) (int, error) { return fmt.Fprint(w, ab.atB(i)) }

// TODO: consider adding a LargeFile wrapper.
// It should read each file once, storing the location of all newlines in each file,
// probably using a compact, delta-based encoding.
// Then Seek/ReadAt to read each line lazily as needed, relying on the OS page cache for performance.
// This will allow diffing giant files with low memory use, at a significant time cost.
// An alternative is to mmap the files, although this is OS-specific and can be fiddly.

// TODO: consider adding a StringIntern type, something like:
//
// type StringIntern struct {
// 	s map[string]*string
// }
//
// func (i *StringIntern) Bytes(b []byte) *string
// func (i *StringIntern) String(s string) *string
//
// And document what it is and why to use it.
// And consider adding helper functions to Strings and Bytes to use it.
// The reason to use it is that a lot of the execution time in diffing
// (which is an expensive operation) is taken up doing string comparisons.
// If you have paid the O(n) cost to intern all strings involved in both A and B,
// then string comparisons are reduced to cheap pointer comparisons.

// TODO: consider adding an "it just works" test helper that accepts two slices (via interface{}),
// diffs them using Strings or Bytes or Slices (using reflect.DeepEqual) as appropriate,
// and calls t.Errorf with a generated diff if they're not equal.
