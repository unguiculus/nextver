package version

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		v       string
		want    *Semver
		wantErr bool
	}{
		{
			name: "default",
			v:    "1.2.3",
			want: &Semver{
				major: "1",
				minor: "2",
				patch: "3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.v, true)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSemver_ApplyPattern(t *testing.T) {
	type args struct {
		current          string
		incrementPattern string
	}
	tests := []struct {
		name           string
		args           args
		want           string
		wantParseErr   bool
		wantApplyError bool
	}{
		{
			name: "reject non semver",
			args: args{
				current:          "0.0",
				incrementPattern: "0.x.0",
			},
			want:         "",
			wantParseErr: true,
		},
		{
			name: "reject empty string",
			args: args{
				current:          "",
				incrementPattern: "",
			},
			want:         "",
			wantParseErr: true,
		},
		{
			name: "reject blank string",
			args: args{
				current:          " ",
				incrementPattern: " ",
			},
			want:         "",
			wantParseErr: true,
		},
		{
			name: "reject non semver increment",
			args: args{
				current:          "0.0.0",
				incrementPattern: "0.x",
			},
			want:           "",
			wantApplyError: true,
		},
		{
			name: "increment major",
			args: args{
				current:          "0.2.4",
				incrementPattern: "x.0.0",
			},
			want: "1.0.0",
		},
		{
			name: "increment minor",
			args: args{
				current:          "0.0.0",
				incrementPattern: "0.x.0",
			},
			want: "0.1.0",
		},
		{
			name: "explicit bump",
			args: args{
				current:          "0.1.0",
				incrementPattern: "1.x.0",
			},
			want: "1.0.0",
		},
		{
			name: "increment patch",
			args: args{
				current:          "0.0.0",
				incrementPattern: "0.0.x",
			},
			want: "0.0.1",
		},
		{
			name: "too many 'x's",
			args: args{
				current:          "0.0.0",
				incrementPattern: "0.x.x",
			},
			want:           "",
			wantApplyError: true,
		},
		{
			name: "too many 'x's, one bordering build",
			args: args{
				current:          "0.0.0",
				incrementPattern: "0.x.x+build",
			},
			want:           "",
			wantApplyError: true,
		},
		{
			name: "too many 'x's, one bordering pre-release",
			args: args{
				current:          "0.0.0",
				incrementPattern: "0.x.x-alpha",
			},
			want:           "",
			wantApplyError: true,
		},
		{
			name: "missing x",
			args: args{
				current:          "0.0.0",
				incrementPattern: "0.0.0",
			},
			want:           "0.0.0",
			wantApplyError: true,
		},
		{
			name: "missing x with x in build",
			args: args{
				current:          "0.0.0",
				incrementPattern: "0.0.0+build.x",
			},
			want: "0.0.0+build.x",
		},
		{
			name: "starting pre-release",
			args: args{
				current:          "0.0.0",
				incrementPattern: "0.0.x-alpha.x",
			},
			want: "0.0.1-alpha.x",
		},
		{
			name: "pre-release and build",
			args: args{
				current:          "0.0.0-alpha+build",
				incrementPattern: "0.0.x-alpha+build",
			},
			want: "0.0.1-alpha+build",
		},
		{
			name: "x only in pre-release",
			args: args{
				current:          "0.0.1-alpha.0",
				incrementPattern: "0.0.1-alpha.x",
			},
			want: "0.0.1-alpha.x",
		},
		{
			name: "incrementing and changing pre-release",
			args: args{
				current:          "0.0.0-alpha.0",
				incrementPattern: "0.0.x-alpha.1",
			},
			want: "0.0.1-alpha.1",
		},
		{
			name: "ignore extra x in pre-release",
			args: args{
				current:          "0.0.0-alpha.0",
				incrementPattern: "0.0.x-alpha.x",
			},
			want: "0.0.1-alpha.x",
		},
		{
			name: "handle build",
			args: args{
				current:          "0.0.0+build",
				incrementPattern: "0.0.x+build",
			},
			want: "0.0.1+build",
		},
		{
			name: "ignore extra x in build",
			args: args{
				current:          "0.0.0+build.x",
				incrementPattern: "0.0.x+build.x",
			},
			want: "0.0.1+build.x",
		},
		{
			name: "ignore x in build",
			args: args{
				current:          "0.0.0+build.x",
				incrementPattern: "0.0.0+build.x",
			},
			want: "0.0.0+build.x",
		},
		{
			name: "major neither a number nor x",
			args: args{
				current:          "0.0.0",
				incrementPattern: "y.0.0",
			},
			want:         "",
			wantParseErr: true,
		},
		{
			name: "increment major with minor and patch set",
			args: args{
				current:          "1.5.2",
				incrementPattern: "x.0.0",
			},
			want:         "2.0.0",
			wantParseErr: false,
		},
		{
			name: "reject decreasing major version when incrementing minor version",
			args: args{
				current:          "1.2.3",
				incrementPattern: "0.x.0",
			},
			want:           "",
			wantApplyError: true,
		},
		{
			name: "reject decreasing major version when incrementing patch version",
			args: args{
				current:          "1.2.3",
				incrementPattern: "0.2.x",
			},
			want:           "",
			wantApplyError: true,
		},
		{
			name: "reject decreasing minor version when incrementing patch version",
			args: args{
				current:          "1.2.3",
				incrementPattern: "1.0.x",
			},
			want:           "",
			wantApplyError: true,
		},
		{
			name: "allow decreasing minor version when incrementing patch version and increasing major version",
			args: args{
				current:          "1.2.3",
				incrementPattern: "2.0.x",
			},
			want: "2.0.0",
		},
		{
			name: "increment minor not knowing major and patch",
			args: args{
				current:          "1.5.2",
				incrementPattern: "?.x.0",
			},
			want: "1.6.0",
		},
		{
			name: "increment patch not knowing major and minor",
			args: args{
				current:          "1.5.2",
				incrementPattern: "?.?.x",
			},
			want: "1.5.3",
		},
		{
			name: "minor neither a number nor x",
			args: args{
				current:          "0.0.0",
				incrementPattern: "0.y.0",
			},
			want:           "",
			wantApplyError: true,
		},
		{
			name: "patch neither a number nor x",
			args: args{
				current:          "0.0.0",
				incrementPattern: "0.0.y",
			},
			want:           "",
			wantApplyError: true,
		},
		{
			name: "support v prefix",
			args: args{
				current:          "v1.0.0",
				incrementPattern: "v1.x.0",
			},
			want: "v1.1.0",
		},
		{
			name: "add v prefix",
			args: args{
				current:          "1.0.0",
				incrementPattern: "v1.x.0",
			},
			want: "v1.1.0",
		},
		{
			name: "remove v prefix",
			args: args{
				current:          "v1.0.0",
				incrementPattern: "1.x.0",
			},
			want: "1.1.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.args.current, true)
			if tt.wantParseErr {
				assert.Error(t, err)
				return
			}
			semver, err := got.ApplyPattern(tt.args.incrementPattern)
			if tt.wantApplyError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, semver.String())
		})
	}
}
