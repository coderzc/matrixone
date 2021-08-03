package or

var (
	SelOr func([]int64, []int64, []int64) int64
)

func selOr(xs, ys, rs []int64) []int64 {
	rs = rs[:0]
	i, j, n, m := 0, 0, len(xs), len(ys)
	for i < n && j < m {
		switch {
		case xs[i] > ys[j]:
			rs = append(rs, ys[j])
			j++
		case xs[i] < ys[j]:
			rs = append(rs, xs[i])
			i++
		default:
			rs = append(rs, xs[i])
			i++
			j++
		}
	}
	for ; i < n; i++ {
		rs = append(rs, xs[i])
	}
	for ; j < m; j++ {
		rs = append(rs, ys[j])
	}
	return rs
}
