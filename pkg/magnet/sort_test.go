package magnet_test

import (
	"github.com/nenadstojanovikj/couch/pkg/magnet"
	"github.com/nenadstojanovikj/couch/pkg/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSortQuality(t *testing.T) {
	testCases := []struct {
		magnets       []storage.Magnet
		expectedOrder []storage.Magnet
		desc          bool
	}{
		{
			magnets: []storage.Magnet{
				{Quality: storage.Quality4K},
				{Quality: storage.QualityHD},
				{Quality: storage.QualityFHD},
				{Quality: storage.QualitySD},
			},
			expectedOrder: []storage.Magnet{
				{Quality: storage.QualitySD},
				{Quality: storage.QualityHD},
				{Quality: storage.QualityFHD},
				{Quality: storage.Quality4K},
			},
			desc: false,
		},
		{
			magnets: []storage.Magnet{
				{Quality: storage.Quality4K},
				{Quality: storage.QualityHD},
				{Quality: storage.QualityFHD},
				{Quality: storage.QualitySD},
			},
			expectedOrder: []storage.Magnet{
				{Quality: storage.Quality4K},
				{Quality: storage.QualityFHD},
				{Quality: storage.QualityHD},
				{Quality: storage.QualitySD},
			},
			desc: true,
		},
	}

	for _, test := range testCases {
		f := magnet.SortQuality(test.desc)
		f(test.magnets)
		assert.Equal(t, test.expectedOrder, test.magnets)
	}
}

func TestSortEncoding(t *testing.T) {
	testCases := []struct {
		magnets       []storage.Magnet
		expectedOrder []storage.Magnet
		desc          bool
	}{
		{
			magnets: []storage.Magnet{
				{Encoding: storage.EncodingHEVC},
				{Encoding: storage.EncodingXVID},
				{Encoding: storage.EncodingVC1},
				{Encoding: storage.Encodingx265},
				{Encoding: storage.Encodingx264},
			},
			expectedOrder: []storage.Magnet{
				{Encoding: storage.EncodingHEVC},
				{Encoding: storage.Encodingx265},
				{Encoding: storage.Encodingx264},
				{Encoding: storage.EncodingVC1},
				{Encoding: storage.EncodingXVID},
			},
			desc: true,
		},
		{
			magnets: []storage.Magnet{
				{Encoding: storage.EncodingHEVC},
				{Encoding: storage.EncodingXVID},
				{Encoding: storage.EncodingVC1},
				{Encoding: storage.Encodingx265},
				{Encoding: storage.Encodingx264},
			},
			expectedOrder: []storage.Magnet{
				{Encoding: storage.EncodingXVID},
				{Encoding: storage.EncodingVC1},
				{Encoding: storage.Encodingx264},
				{Encoding: storage.Encodingx265},
				{Encoding: storage.EncodingHEVC},
			},
			desc: false,
		},
	}

	for _, test := range testCases {
		f := magnet.SortEncoding(test.desc)
		f(test.magnets)
		assert.Equal(t, test.expectedOrder, test.magnets)
	}
}
