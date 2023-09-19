package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertFromLocalityStr(t *testing.T) {
	locality := "FULL{1}@zone1, FULL{1}@zone2, FULL{1}@zone3"
	replicas := ConvertFromLocalityStr(locality)
	require.Equal(t, 3, len(replicas))
}
