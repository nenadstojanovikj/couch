package magnet_test

import (
	"fmt"
	"testing"

	"github.com/nenad/couch/pkg/magnet"
	"github.com/nenad/couch/pkg/storage"
	"github.com/stretchr/testify/assert"
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
				{Encoding: storage.Encodingx265},
				{Encoding: storage.EncodingXVID},
				{Encoding: storage.EncodingVC1},
				{Encoding: storage.Encodingx265},
				{Encoding: storage.Encodingx264},
			},
			expectedOrder: []storage.Magnet{
				{Encoding: storage.Encodingx265},
				{Encoding: storage.Encodingx265},
				{Encoding: storage.Encodingx264},
				{Encoding: storage.EncodingVC1},
				{Encoding: storage.EncodingXVID},
			},
			desc: true,
		},
		{
			magnets: []storage.Magnet{
				{Encoding: storage.Encodingx265},
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
				{Encoding: storage.Encodingx265},
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

func TestSortSize(t *testing.T) {
	testCases := []struct {
		magnets       []storage.Magnet
		expectedOrder []storage.Magnet
		desc          bool
	}{
		{
			magnets: []storage.Magnet{
				{Size: 1},
				{Size: 3},
				{Size: 2},
			},
			expectedOrder: []storage.Magnet{
				{Size: 3},
				{Size: 2},
				{Size: 1},
			},
			desc: true,
		},
		{
			magnets: []storage.Magnet{
				{Size: 1},
				{Size: 3},
				{Size: 2},
			},
			expectedOrder: []storage.Magnet{
				{Size: 1},
				{Size: 2},
				{Size: 3},
			},
			desc: false,
		},
	}

	for _, test := range testCases {
		f := magnet.SortSize(test.desc)
		f(test.magnets)
		assert.Equal(t, test.expectedOrder, test.magnets)
	}
}

func TestSortCombined(t *testing.T) {
	testCases := []struct {
		magnets       []storage.Magnet
		expectedOrder []storage.Magnet
	}{
		{
			magnets: []storage.Magnet{
				{Size: 9, Encoding: storage.Encodingx265, Quality: storage.Quality4K},
				{Size: 9, Encoding: storage.Encodingx265, Quality: storage.QualityFHD},
				{Size: 8, Encoding: storage.Encodingx265, Quality: storage.QualityFHD},
				{Size: 8, Encoding: storage.Encodingx265, Quality: storage.Quality4K},
				{Size: 7, Encoding: storage.Encodingx264, Quality: storage.Quality4K},
				{Size: 4, Encoding: storage.Encodingx264, Quality: storage.Quality4K},
				{Size: 10, Encoding: storage.Encodingx265, Quality: storage.Quality4K},
				{Size: 3, Encoding: storage.Encodingx264, Quality: storage.Quality4K},
			},
			expectedOrder: []storage.Magnet{
				{Size: 8, Encoding: storage.Encodingx265, Quality: storage.Quality4K},
				{Size: 9, Encoding: storage.Encodingx265, Quality: storage.Quality4K},
				{Size: 10, Encoding: storage.Encodingx265, Quality: storage.Quality4K},
				{Size: 3, Encoding: storage.Encodingx264, Quality: storage.Quality4K},
				{Size: 4, Encoding: storage.Encodingx264, Quality: storage.Quality4K},
				{Size: 7, Encoding: storage.Encodingx264, Quality: storage.Quality4K},
				{Size: 8, Encoding: storage.Encodingx265, Quality: storage.QualityFHD},
				{Size: 9, Encoding: storage.Encodingx265, Quality: storage.QualityFHD},
			},
		},
	}

	for _, test := range testCases {
		fs := []magnet.ProcessFunc{
			magnet.SortSize(false),
			magnet.SortEncoding(true),
			magnet.SortQuality(true),
		}
		for i, f := range fs {
			fmt.Printf("Test %d\n", i)
			fmt.Printf("Before:\n")
			for i, m := range test.magnets {
				fmt.Printf("%d: %d %s %s\n", i, m.Size, m.Encoding, m.Quality)
			}
			f(test.magnets)
			fmt.Printf("After: \n")
			for i, m := range test.magnets {
				fmt.Printf("%d: %d %s %s\n", i, m.Size, m.Encoding, m.Quality)
			}
		}
		assert.Equal(t, test.expectedOrder, test.magnets)
	}
}
