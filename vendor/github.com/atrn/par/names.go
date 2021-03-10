package par

// A small "abuse" using import . allows par to inject other names for
// its functions.  The forms prefixed with "par." read well however
// when the package is used via an "import ," it defines and promotes
// the names below.
//
var (
	PAR       = DO
	PAR_FOR   = FOR
	PARfn     = DOfn
	PAR_FORfn = FORfn
)
