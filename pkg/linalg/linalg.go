package linalg

import (
	"runtime"
)

func PageRank(graph map[string]map[string]int) map[string]float64 {
	// n - amount of vertices
	n := len(graph)

	idToIndex := make(map[string]int, n)
	indexToId := make(map[int]string, n)
	index := 0
	for id := range graph {
		idToIndex[id] = index
		indexToId[index] = id
		index++
	}

	// A - transition matrix
	A := make([][]float64, n)
	for i := 0; i < n; i++ {
		A[i] = make([]float64, n)
	}
	for id1 := range graph {
		for id2 := range graph[id1] {
			from := idToIndex[id1]
			to := idToIndex[id2]
			A[to][from]++
		}
	}
	// s - column sum
	s := make([]float64, n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			s[i] += A[j][i]
		}
	}
	for i := 0; i < n; i++ {
		if s[i] != 0 {
			for j := 0; j < n; j++ {
				A[j][i] /= s[i]
			}
		}
	}

	// B - column-stochastic matrix
	B := make([][]float64, n)
	for i := 0; i < n; i++ {
		B[i] = make([]float64, n)
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			B[i][j] = 1 / float64(n)
		}
	}

	// p - damping factor
	p := 0.15

	// M - PageRank matrix
	M := make([][]float64, n)
	for i := 0; i < n; i++ {
		M[i] = make([]float64, n)
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			M[i][j] = (1-p)*A[i][j] + p*B[i][j]
		}
	}

	// v - significance vector
	v := make([]float64, n)
	for i := 0; i < n; i++ {
		v[i] = 1 / float64(n)
	}
	for i := 0; i < 50; i++ {
		vv := make([]float64, n)
		for j := 0; j < n; j++ {
			for k := 0; k < n; k++ {
				vv[j] += M[j][k] * v[k]
			}
		}
		sum := 0.0
		for j := 0; j < len(vv); j++ {
			sum += vv[j]
		}
		for j := 0; j < len(vv); j++ {
			vv[j] /= sum
		}
		v = vv
		runtime.Gosched()
	}

	res := make(map[string]float64, n)
	for index, rating := range v {
		id := indexToId[index]
		res[id] = rating
	}
	return res
}
