package utils

import (
	"fmt"
	"testing"

	"github.com/codingsince1985/geo-golang/openstreetmap"
	"github.com/stretchr/testify/require"
)

type pair struct {
	Address Location
	Result  []string
}

func TestPrettyAddress(t *testing.T) {
	expected := []pair{
		{
			Address: Location{
				Latitude:  -39.6375091,
				Longitude: 175.2222849,
			},
			Result: []string{"Matahiwi Manawatū-Whanganui New Zealand _ Aotearoa", "New Zealand _ Aotearoa", "Matahiwi Track, Matahiwi, Whanganui District, Manawatū-Whanganui, New Zealand _ Aotearoa"},
		},
		{
			Address: Location{
				Latitude:  52.547567,
				Longitude: 13.385176,
			},
			Result: []string{"Berlin Deutschland", "Deutschland", "Flakturm Humboldthain, Humboldtsteg, Gesundbrunnen, Mitte, Berlin, 13357, Deutschland"},
		},
	}

	for valueIndex, value := range expected {
		t.Run(fmt.Sprintf("Pair %d", valueIndex), func(t *testing.T) {
			service := openstreetmap.Geocoder()
			address, err := service.ReverseGeocode(value.Address.Latitude, value.Address.Longitude)
			require.NoError(t, err)

			resp := getPrettyAddress(format1{}, address)
			require.Equal(t, value.Result[0], resp)

			resp = getPrettyAddress(format2{}, address)
			require.Equal(t, value.Result[1], resp)

			resp = getPrettyAddress(format3{}, address)
			require.Equal(t, value.Result[2], resp)
		})
	}
}
