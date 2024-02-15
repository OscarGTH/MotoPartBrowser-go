package main

import "testing"

func Test_extractYear(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"Test plain year", args{"2019"}, 2019},
		{"Test year with text", args{"Suzuki 2019"}, 2019},
		{"Test year with text and spaces", args{"Suzuki RX 2019"}, 2019},
		{"Test year with text and spaces and other numbers", args{"Suzuki RX 203 2019"}, 2019},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractYear(tt.args.s); got != tt.want {
				t.Errorf("extractYear() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractBrand(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Test plain brand", args{"Suzuki"}, "Suzuki"},
		{"Test brand with text", args{"Suzuki RX"}, "Suzuki"},
		{"Test brand with text and spaces", args{"Suzuki RX 203"}, "Suzuki"},
		{"Test brand with text and spaces and other numbers", args{"Suzuki RX 203 2019"}, "Suzuki"},
		{"Test brand with two brands in the name", args{"Suzuki RX 203 2019 or Kawasaki XP"}, "Suzuki"},
		{"Test multipart name with spaces as separators", args{"Harley Davidson RX203 2020"}, "Harley Davidson"},
		{"Test multipart name with dashes as separators", args{"Harley-Davidson RX203 2020"}, "Harley Davidson"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractBrand(tt.args.s, BrandReMatcher); got != tt.want {
				t.Errorf("extractBrand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractModel(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Test model with brand", args{"Suzuki RX"}, "RX"},
		{"Test model with brand and year", args{"Suzuki RX 2019"}, "RX"},
		{"Test model with spaces and with brand and year", args{"Suzuki RX 3 2019"}, "RX 3"},
		{"Test model with spaces and with brand and year and other numbers", args{"Suzuki RX 3 2019 203"}, "RX 3"},
		{"Test model with dashes in model name", args{"Suzuki RX-3 2019"}, "RX-3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractModel(tt.args.s); got != tt.want {
				t.Errorf("extractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractVehicleIdentifier(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Test normal link", args{"https://www.purkuosat.net/apriliamx12504.htm"}, "apriliamx12504"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractVehicleIdentifier(tt.args.s); got != tt.want {
				t.Errorf("extractVehicleIdentifier() = %v, want %v", got, tt.want)
			}
		})
	}
}
