package stringnorm

// NormList unpacks trees of normalizers into a simple List of atomic
// Normalizers.
func NormList(norm Normalizer) List {
	res := List{}
	var traverse func(Normalizer)
	traverse = func(n Normalizer) {
		switch act := n.(type) {
		case List:
			for _, child := range act {
				traverse(child)
			}
		default:
			res = append(res, n)
		}
	}
	traverse(norm)
	return res
}
