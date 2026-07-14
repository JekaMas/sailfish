package sailfish

// Error is an allocation-free, comparable package error.
//
// Exported errors are typed string constants. They work with errors.Is when
// returned directly or wrapped with fmt.Errorf and %w.
type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrSyntax               Error = "sailfish: invalid syntax"
	ErrRange                Error = "sailfish: value does not fit unit type"
	ErrPrecision            Error = "sailfish: too many fractional digits"
	ErrScale                Error = "sailfish: scale is unsupported by unit type"
	ErrOverflow             Error = "sailfish: addition overflow"
	ErrUnderflow            Error = "sailfish: subtraction underflow"
	ErrUninitializedCodec   Error = "sailfish: uninitialized codec"
	ErrNilDestination       Error = "sailfish: nil destination"
	ErrCBORSyntax           Error = "sailfish: invalid CBOR"
	ErrCBORNonDeterministic Error = "sailfish: non-deterministic CBOR"
)

// Pre-box the fixed errors once. Returning an Error directly as an error
// interface from generic code otherwise allocates a string header on each
// failure. The exported source of truth remains the typed string constants.
var (
	boxedErrSyntax               error = ErrSyntax
	boxedErrRange                error = ErrRange
	boxedErrPrecision            error = ErrPrecision
	boxedErrScale                error = ErrScale
	boxedErrOverflow             error = ErrOverflow
	boxedErrUnderflow            error = ErrUnderflow
	boxedErrUninitializedCodec   error = ErrUninitializedCodec
	boxedErrNilDestination       error = ErrNilDestination
	boxedErrCBORSyntax           error = ErrCBORSyntax
	boxedErrCBORNonDeterministic error = ErrCBORNonDeterministic
)

func boxedError(err Error) error {
	switch err {
	case "":
		return nil
	case ErrSyntax:
		return boxedErrSyntax
	case ErrRange:
		return boxedErrRange
	case ErrPrecision:
		return boxedErrPrecision
	case ErrScale:
		return boxedErrScale
	case ErrOverflow:
		return boxedErrOverflow
	case ErrUnderflow:
		return boxedErrUnderflow
	case ErrUninitializedCodec:
		return boxedErrUninitializedCodec
	case ErrNilDestination:
		return boxedErrNilDestination
	case ErrCBORSyntax:
		return boxedErrCBORSyntax
	case ErrCBORNonDeterministic:
		return boxedErrCBORNonDeterministic
	default:
		// Internal callers only produce the constants above. Preserve correct
		// behavior for a future Error value even if it takes the cold allocation.
		return err
	}
}
