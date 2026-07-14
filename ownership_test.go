package sailfish

import (
	"sync"
	"testing"

	"github.com/holiman/uint256"
)

func TestCanonicalRetentionAndCompactOwnership(t *testing.T) {
	t.Parallel()

	canonicalInput := "123.31232"
	canonical, err := New[PriceScale5](canonicalInput)
	if err != nil || !canonical.HasRepresentation() || canonical.String() != canonicalInput {
		t.Fatalf("canonical = %v %#v", err, canonical)
	}

	nonCanonical, err := New[PriceScale5]("00123.31")
	if err != nil || nonCanonical.HasRepresentation() || nonCanonical.String() != "123.31000" {
		t.Fatalf("non-canonical = %v %q", err, nonCanonical.String())
	}

	compact, err := NewCompact[PriceScale5](canonicalInput)
	if err != nil || compact.HasRepresentation() {
		t.Fatalf("compact retained input: %v", err)
	}

	bytesInput := []byte(canonicalInput)
	fromBytes, err := NewBytes[PriceScale5](bytesInput)
	if err != nil || fromBytes.HasRepresentation() {
		t.Fatalf("bytes retained input: %v", err)
	}
	bytesInput[0] = '9'
	if fromBytes.String() != canonicalInput {
		t.Fatalf("byte mutation changed value: %q", fromBytes.String())
	}
}

func TestRepresentationInvalidationDoesNotMutateReturnedString(t *testing.T) {
	t.Parallel()

	value, _ := New[PriceScale5]("9.99999")
	delta, _ := New[PriceScale5]("0.00001")
	old := value.String()

	if value.AddAssign(delta) {
		t.Fatal("unexpected overflow")
	}
	if old != "9.99999" {
		t.Fatalf("returned string mutated: %q", old)
	}
	if value.HasRepresentation() {
		t.Fatal("value-changing mutation preserved stale representation")
	}
	if got := value.String(); got != "10.00000" {
		t.Fatalf("mutated value = %q", got)
	}
}

func TestNoOpMutationPreservesRepresentation(t *testing.T) {
	t.Parallel()

	value, _ := New[PriceScale5]("1.20000")
	zero, _ := New[PriceScale5]("0.00000")
	if value.AddAssign(zero) || !value.HasRepresentation() {
		t.Fatal("adding zero invalidated representation")
	}
	value.SetUnits(value.Units())
	if !value.HasRepresentation() {
		t.Fatal("setting identical units invalidated representation")
	}
}

func TestCanonicalReturnsIndependentCopy(t *testing.T) {
	t.Parallel()

	codec := MustCodec[PriceScale5]()
	original := codec.FromUnits(12_000_000)
	canonical := original.Canonical()
	if original.HasRepresentation() || !canonical.HasRepresentation() {
		t.Fatalf("original=%v canonical=%v", original.HasRepresentation(), canonical.HasRepresentation())
	}
}

func TestConcurrentStringReadsHaveNoMutableCache(t *testing.T) {
	t.Parallel()

	value, _ := NewFromUnits[uint256Scale18](uint256.Int{1, 2, 3, 4})
	const readers = 16
	var wg sync.WaitGroup
	wg.Add(readers)
	for range readers {
		go func() {
			defer wg.Done()
			for range 1_000 {
				if value.String() == "" {
					t.Error("empty string")
				}
			}
		}()
	}
	wg.Wait()
}
