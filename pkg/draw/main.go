package draw

import (
	"math"
	"math/big"
	"sort"
	"time"

	tm "github.com/buger/goterm"
	"gonum.org/v1/gonum/stat"
)

var (
	threshold = big.NewFloat(0.00001)
)

type data struct {
	time  time.Time
	value float64
}

type Draw struct {
	data   []data
	sorted []float64
}

func (d *Draw) Append(time time.Time, value float64) {
	d.data = append(d.data, data{time, value})

	sorted := make([]float64, len(d.data))
	for idx, v := range d.data {
		sorted[idx] = v.value
	}
	sort.Float64s(sorted)

	d.sorted = sorted
}

func (d *Draw) Flush() error {
	if len(d.data) == 0 {
		return nil
	}

	tm.Clear()

	// Write it at the top of the screen
	tm.MoveCursor(1, 1)

	dt := new(tm.DataTable)
	dt.AddColumn("Time")
	dt.AddColumn("Power W")

	anmlDT := new(tm.DataTable)
	anmlDT.AddColumn("Time")
	anmlDT.AddColumn("Anomaly")

	last := len(d.data) - 20
	if last < 0 {
		last = 0
	}
	for _, v := range d.data[last:] {
		var anomaly float64

		if d.isAnomaly(v.value) {
			anomaly = 1
		}

		dt.AddRow(float64(v.time.Unix()), v.value)
		anmlDT.AddRow(float64(v.time.Unix()), anomaly)
	}

	chart := tm.NewLineChart(150, 30)
	if _, err := tm.Println(chart.Draw(dt)); err != nil {
		return err
	}

	anmlChart := tm.NewLineChart(150, 30)
	if _, err := tm.Println(anmlChart.Draw(anmlDT)); err != nil {
		return err
	}

	mean := stat.Mean(d.sorted, nil)
	variance := stat.Variance(d.sorted, nil)
	stddev := math.Sqrt(variance)
	median := stat.Quantile(0.5, stat.Empirical, d.sorted, nil)
	tm.Printf("items     %d\n", len(d.data))
	tm.Printf("mean      %v\n", mean)
	tm.Printf("median    %v\n", median)
	tm.Printf("std-dev   %v\n", stddev)
	tm.Printf("variance  %v\n", variance)

	tm.Flush()

	return nil
}

func (d *Draw) isAnomaly(v float64) bool {
	if len(d.sorted) == 0 {
		return false
	}

	sorted := d.sorted

	quartile1 := median(sorted[:len(sorted)/2])
	quartile3 := median(sorted[len(sorted)/2:])

	iqr := (quartile3 - quartile1)

	low := quartile1 - (1.9 * iqr)
	high := quartile3 + (1.9 * iqr)

	return v < low || v > high
}

func median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}

	if len(xs) == 1 {
		return xs[0]
	}
	med := xs[len(xs)/2]
	if len(xs)%2 == 0 {
		med += xs[len(xs)/2-1]
		med /= 2
	}
	return med
}
