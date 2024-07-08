package maths

// Pretty much transcribed from https://stackoverflow.com/a/147539
// LcmMultiple returns the greatest common multiple of a and b...

func Gcd(a, b int) int {
	// Return greatest common divisor using Euclid's Algorithm.
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func Lcm(a, b int) int {
	// Return lowest common multiple.
	return a * b / Gcd(a, b)
}

func LcmMultiple(args ...int) int {
	// """Return lcm of args."""
	if len(args) == 0 {
		return 1
	}
	rv := args[0]
	for i := 1; i < len(args); i++ {
		rv = Lcm(rv, args[i])
	}
	return rv
}
