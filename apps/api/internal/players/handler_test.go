package players

import "testing"

func TestComputeBoundsEmptyPositions(t *testing.T) {
	bounds := computeBounds([]Point{})
	if bounds != (WorldBounds{}) {
		t.Fatalf("expected zero bounds, got %+v", bounds)
	}
}

func TestComputeBoundsAppliesMinimumPadding(t *testing.T) {
	positions := []Point{{X: 100, Y: 100}, {X: 100, Y: 100}}
	bounds := computeBounds(positions)

	if bounds.MinX != 100-minBoundsPadding || bounds.MaxX != 100+minBoundsPadding {
		t.Fatalf("expected minimum padding on x, got %+v", bounds)
	}
	if bounds.MinY != 100-minBoundsPadding || bounds.MaxY != 100+minBoundsPadding {
		t.Fatalf("expected minimum padding on y, got %+v", bounds)
	}
}

func TestComputeBoundsAppliesRatioPaddingForLargeRanges(t *testing.T) {
	positions := []Point{{X: 0, Y: 0}, {X: 10000, Y: 2000}}
	bounds := computeBounds(positions)

	expectedPadX := boundsPaddingRatio * 10000
	expectedPadY := boundsPaddingRatio * 2000

	if bounds.MinX != -expectedPadX || bounds.MaxX != 10000+expectedPadX {
		t.Fatalf("expected ratio padding on x, got %+v", bounds)
	}
	if bounds.MinY != -expectedPadY || bounds.MaxY != 2000+expectedPadY {
		t.Fatalf("expected ratio padding on y, got %+v", bounds)
	}
}

func TestComputeBoundsTracksActualMinMax(t *testing.T) {
	positions := []Point{
		{X: 50, Y: -20},
		{X: -30, Y: 80},
		{X: 10, Y: 10},
	}
	bounds := computeBounds(positions)

	if bounds.MinX != -30-minBoundsPadding {
		t.Fatalf("expected minX -30 with padding, got %v", bounds.MinX)
	}
	if bounds.MaxX != 50+minBoundsPadding {
		t.Fatalf("expected maxX 50 with padding, got %v", bounds.MaxX)
	}
	if bounds.MinY != -20-minBoundsPadding {
		t.Fatalf("expected minY -20 with padding, got %v", bounds.MinY)
	}
	if bounds.MaxY != 80+minBoundsPadding {
		t.Fatalf("expected maxY 80 with padding, got %v", bounds.MaxY)
	}
}
