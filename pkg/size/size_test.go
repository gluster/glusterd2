package size

import "testing"

var (
	kbSizeTests = []struct {
		s    Size
		want float64
	}{
		{Size(1000), 1},
		{Size(2000), 2},
		{Size(2500), 2.5},
		{Size(8750), 8.75},
	}

	mbSizeTests = []struct {
		s    Size
		want float64
	}{
		{Size(1e6), 1},
		{Size(2.5e6), 2.5},
		{Size(8.75e6), 8.75},
		{Size(2047e6), 2047},
	}

	gbSizeTests = []struct {
		s    Size
		want float64
	}{
		{Size(1e9), 1},
		{Size(2.5e9), 2.5},
		{Size(8.75e9), 8.75},
		{Size(0.25e9), 0.25},
	}

	tbSizeTests = []struct {
		s    Size
		want float64
	}{
		{Size(1e12), 1},
		{Size(2.5e12), 2.5},
		{Size(8.75e12), 8.75},
		{Size(0.75e12), 0.75},
	}

	stringSizeTests = []struct {
		s    Size
		want string
	}{
		{Size(1099511627776), "1.00TiB"},
		{Size(2684354560), "2.50GiB"},
		{Size(9175040), "8.75MiB"},
		{Size(786432), "768.00KiB"},
		{Size(500), "500 B"},
	}

	parseSizeTests = []struct {
		s    string
		want Size
	}{
		// Byte Format
		{"1B", Byte},
		{"512.0 B", 512 * Byte},
		{"125.0B", 125 * Byte},

		// Binary format
		{"1GiB", Size(1 * float64(GiB))},
		{"2.5GiB", Size(2.5 * float64(GiB))},
		{"1MiB", Size(1 * float64(MiB))},
		{"100.5MiB", Size(100.5 * float64(MiB))},
		{"50KiB", Size(50 * float64(KiB))},
		{"0050KiB", Size(50 * float64(KiB))},
		{"2.50KiB", Size(2.5 * float64(KiB))},
		{"2.50TiB", Size(2.5 * float64(TiB))},

		{"1Gi", Size(1 * float64(GiB))},
		{"2.5Gi", Size(2.5 * float64(GiB))},
		{"1Mi", Size(1 * float64(MiB))},
		{"100.5Mi", Size(100.5 * float64(MiB))},
		{"50Ki", Size(50 * float64(KiB))},
		{"0050Ki", Size(50 * float64(KiB))},
		{"2.50Ki", Size(2.5 * float64(KiB))},
		{"2.50Ti", Size(2.5 * float64(TiB))},

		{"100.5M", Size(100.5 * float64(MiB))},
		{"50K", Size(50 * float64(KiB))},
		{"0050K", Size(50 * float64(KiB))},
		{"2.50K", Size(2.5 * float64(KiB))},
		{"2.50T", Size(2.5 * float64(TiB))},

		// Decimal format
		{"2.50TB", Size(2.5 * float64(TB))},
		{"2.50MB", Size(2.5 * float64(MB))},
		{"0.5KB", Size(0.5 * float64(KB))},
		{"052GB", Size(52 * float64(GB))},

		// having space in between
		{"0.5 KB", Size(0.5 * float64(KB))},
		{"052 GB", Size(52 * float64(GB))},
		{"0050 KiB", Size(50 * float64(KiB))},
		{"2.5 KiB", Size(2.5 * float64(KiB))},
		{"2.50 TiB", Size(2.5 * float64(TiB))},
	}

	parseSizeFailureTest = []string{
		"1xGiB",
		"x1TiB",
		"5kiB",
		"7.4xKiB",
		"7.4KKiB",
		"7.4KMiB",
		"7.4MiBT",
		//
		"5KBM",
		"x5KB",
		"05xMB",
		"5.5.5MB",
		"5BB",
		"4.5.5B",
	}
)

func TestSizeBytes(t *testing.T) {
	var s Size = 2048
	bytes := s.Bytes()
	if bytes != 2048 {
		t.Errorf("s.Bytes() = %v; want: %v", bytes, 2048)
	}
}

func TestSizeKiloBytes(t *testing.T) {
	for _, tt := range kbSizeTests {
		if got := tt.s.KiloBytes(); got != tt.want {
			t.Errorf("s.KiloBytes() = %v; want: %v", got, tt.want)
		}
	}
}

func TestSizeMegaBytes(t *testing.T) {
	for _, tt := range mbSizeTests {
		if got := tt.s.MegaBytes(); got != tt.want {
			t.Errorf("s.MegaBytes() = %v; want: %v", got, tt.want)
		}
	}
}

func TestSizeGigaBytes(t *testing.T) {
	for _, tt := range gbSizeTests {
		if got := tt.s.GigaBytes(); got != tt.want {
			t.Errorf("s.GigaBytes() = %v; want: %v", got, tt.want)
		}
	}
}

func TestSizeTeraBytes(t *testing.T) {
	for _, tt := range tbSizeTests {
		if got := tt.s.TeraBytes(); got != tt.want {
			t.Errorf("s.TeraBytes() = %v; want: %v", got, tt.want)
		}
	}
}

func TestSizeString(t *testing.T) {
	for _, tt := range stringSizeTests {
		if got := tt.s.String(); got != tt.want {
			t.Errorf("s.String() = %v; want: %v", got, tt.want)
		}
	}
}

func TestParse(t *testing.T) {
	for _, tt := range parseSizeTests {
		got, err := Parse(tt.s)
		if err != nil {
			t.Error("error in s.Parse() :", err)
		} else {
			if got != tt.want {
				t.Errorf("s.Parse() = %v; want: %v", got, tt.want)
			}
		}
	}
}

func TestParseFailure(t *testing.T) {
	for _, s := range parseSizeFailureTest {
		if sz, err := Parse(s); err == nil {
			t.Errorf("s.Parse() = %v; wanted error", sz)
		}
	}
}
