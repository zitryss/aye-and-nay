package linalg_test

import (
	"math/rand"
	"strconv"
	"testing"

	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/linalg"
)

func TestPageRank(t *testing.T) {
	edgs := map[string]map[string]int{}
	edgs["Design Patterns"] = map[string]int{}
	edgs["Refactoring"] = map[string]int{}
	edgs["Clean Code"] = map[string]int{}
	edgs["Code Complete"] = map[string]int{}
	edgs["The Progmatic Programmer"] = map[string]int{}
	edgs["Refactoring"]["Design Patterns"]++
	edgs["Clean Code"]["Design Patterns"]++
	edgs["Clean Code"]["Refactoring"]++
	edgs["Code Complete"]["Design Patterns"]++
	edgs["Code Complete"]["Refactoring"]++
	edgs["Code Complete"]["Clean Code"]++
	edgs["The Progmatic Programmer"]["Design Patterns"]++
	edgs["The Progmatic Programmer"]["Refactoring"]++
	edgs["The Progmatic Programmer"]["Clean Code"]++
	edgs["The Progmatic Programmer"]["Code Complete"]++
	got := linalg.PageRank(edgs)
	want := map[string]float64{}
	want["Design Patterns"] = 0.539773357682638
	want["Refactoring"] = 0.20997909420705596
	want["Clean Code"] = 0.11761540730647063
	want["Code Complete"] = 0.07719901505201851
	want["The Progmatic Programmer"] = 0.055433125751816706
	if !EqualMap(got, want) {
		t.Error("!equalMap(got, want)")
	}
}

func BenchmarkPageRank(b *testing.B) {
	for i := 995; i <= 1005; i++ {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			edgs := map[string]map[string]int{}
			for j := 0; j < i; j++ {
				node := strconv.Itoa(j)
				edgs[node] = map[string]int{}
			}
			for j := 0; j < i; j++ {
				from := strconv.Itoa(rand.Intn(i))
				to := strconv.Itoa(rand.Intn(i))
				edgs[from][to]++
			}
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				linalg.PageRank(edgs)
			}
		})
	}
}
