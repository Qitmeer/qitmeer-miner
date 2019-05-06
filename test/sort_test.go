package test

import (
	"testing"
	"sort"
	"github.com/stretchr/testify/assert"
)

type people struct {
	ID int
}

type SortByID []people

func (s SortByID) Len() int {
	return len(s)
}

func (s SortByID)Swap(i,j int)  {
	s[i],s[j] = s[j],s[i]
}

func (s SortByID)Less(i,j int) bool {
	return s[i].ID < s[j].ID
}

func TestSort(t *testing.T) {
	ps := []people{{ID: 2}, {ID: 1}, {ID: 15}, {ID: 13}}
	sort.Sort(SortByID(ps))
	assert.Equal(t, []people{{ID: 1}, {ID: 2}, {ID: 13}, {ID: 15}}, ps, "Not Equal The Result!")
}