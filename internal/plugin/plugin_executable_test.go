package plugin

import (
	"context"
	"testing"

	"github.com/hashicorp/boundary/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/boundary/internal/db"
	"github.com/hashicorp/boundary/internal/plugin/store"
)

func TestPluginExecutable_Create(t *testing.T) {
	conn, _ := db.TestSetup(t, "postgres")
	plg := testPlugin(t, conn, "test")
	plgVer := testPluginVersion(t, conn, plg.GetPublicId(), "0.0.1")

	sample := []byte("this is just an example")

	type args struct {
		verId string
		os    OperatingSystem
		arch  Architecture
		exe   []byte
		opts  []Option
	}

	tests := []struct {
		name    string
		args    args
		want    *PluginExecutable
		wantErr bool
	}{
		{
			name: "blank-versionid",
			args: args{
				os:   WindowsOS,
				arch: Amd64Arch,
				exe:  sample,
			},
			want: &PluginExecutable{
				PluginExecutable: &store.PluginExecutable{
					OperatingSystem: WindowsOS,
					Architecture:    Amd64Arch,
					Executable:      sample,
				},
			},
			wantErr: true,
		},
		{
			name: "blank-os",
			args: args{
				verId: plgVer.GetPublicId(),
				arch:  Amd64Arch,
				exe:   sample,
			},
			want: &PluginExecutable{
				PluginExecutable: &store.PluginExecutable{
					VersionId:    plgVer.GetPublicId(),
					Architecture: Amd64Arch,
					Executable:   sample,
				},
			},
			wantErr: true,
		},
		{
			name: "blank-arch",
			args: args{
				verId: plgVer.GetPublicId(),
				os:    WindowsOS,
				exe:   sample,
			},
			want: &PluginExecutable{
				PluginExecutable: &store.PluginExecutable{
					VersionId:       plgVer.GetPublicId(),
					OperatingSystem: WindowsOS,
					Executable:      sample,
				},
			},
			wantErr: true,
		},
		{
			name: "unknown-os",
			args: args{
				verId: plgVer.GetPublicId(),
				os:    OperatingSystem("something"),
				arch:  Amd64Arch,
				exe:   sample,
			},
			want: &PluginExecutable{
				PluginExecutable: &store.PluginExecutable{
					VersionId:       plgVer.GetPublicId(),
					OperatingSystem: "something",
					Architecture:    Amd64Arch,
					Executable:      sample,
				},
			},
		},
		{
			name: "unknown-arch",
			args: args{
				verId: plgVer.GetPublicId(),
				os:    WindowsOS,
				arch:  Architecture("something"),
				exe:   sample,
			},
			want: &PluginExecutable{
				PluginExecutable: &store.PluginExecutable{
					VersionId:    plgVer.GetPublicId(),
					Architecture: "something",
					Executable:   sample,
				},
			},
		},
		{
			name: "success",
			args: args{
				verId: plgVer.GetPublicId(),
				os:    WindowsOS,
				arch:  Amd64Arch,
				exe:   sample,
			},
			want: &PluginExecutable{
				PluginExecutable: &store.PluginExecutable{
					VersionId:       plgVer.GetPublicId(),
					OperatingSystem: WindowsOS,
					Architecture:    Amd64Arch,
					Executable:      sample,
				},
			},
		},
		// this must be run after success
		{
			name: "duplicate",
			args: args{
				verId: plgVer.GetPublicId(),
				os:    WindowsOS,
				arch:  Amd64Arch,
				exe:   sample,
			},
			want: &PluginExecutable{
				PluginExecutable: &store.PluginExecutable{
					VersionId:       plgVer.GetPublicId(),
					OperatingSystem: WindowsOS,
					Architecture:    Amd64Arch,
					Executable:      sample,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := NewPluginExecutable(tt.args.verId, tt.args.os, tt.args.arch, tt.args.exe)
			require.NotNil(t, got)

			w := db.New(conn)
			err := w.Create(context.Background(), got)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPluginExecutable_Delete(t *testing.T) {
	conn, _ := db.TestSetup(t, "postgres")
	w := db.New(conn)
	ctx := context.Background()
	exe := []byte("test")

	t.Run("cascade-plugin", func(t *testing.T) {
		plg := testPlugin(t, conn, "cascade-plugin")
		plgVer := testPluginVersion(t, conn, plg.GetPublicId(), "0.0.1")

		plgExe := NewPluginExecutable(plgVer.GetPublicId(), WindowsOS, Amd64Arch, exe)
		require.NoError(t, w.Create(ctx, plgExe))

		deleted, err := w.Delete(ctx, plg)
		require.NoError(t, err)
		require.Equal(t, 1, deleted)

		err = w.LookupWhere(ctx, plgExe, "version_id=?", plgVer.GetPublicId())
		require.Error(t, err)
		assert.True(t, errors.IsNotFoundError(err))
	})

	t.Run("cascade-version", func(t *testing.T) {
		plg := testPlugin(t, conn, "cascade-version")
		plgVer := testPluginVersion(t, conn, plg.GetPublicId(), "0.0.1")

		plgExe := NewPluginExecutable(plgVer.GetPublicId(), WindowsOS, Amd64Arch, exe)
		require.NoError(t, w.Create(ctx, plgExe))

		deleted, err := w.Delete(ctx, plgVer)
		require.NoError(t, err)
		require.Equal(t, 1, deleted)

		err = w.LookupWhere(ctx, plgExe, "version_id=?", plgVer.GetPublicId())
		require.Error(t, err)
		assert.True(t, errors.IsNotFoundError(err))
	})

	t.Run("direct-delete", func(t *testing.T) {
		plg := testPlugin(t, conn, "direct-delete")
		plgVer := testPluginVersion(t, conn, plg.GetPublicId(), "0.0.1")

		plgExe := NewPluginExecutable(plgVer.GetPublicId(), WindowsOS, Amd64Arch, exe)
		require.NoError(t, w.Create(ctx, plgExe))

		deleted, err := w.Delete(ctx, plgExe)
		require.NoError(t, err)
		require.Equal(t, 1, deleted)

		err = w.LookupWhere(ctx, plgExe, "version_id=?", plgVer.GetPublicId())
		require.Error(t, err)
		assert.True(t, errors.IsNotFoundError(err))
	})
}

func TestPluginExecutable_SetTableName(t *testing.T) {
	defaultTableName := "plugin_executable"
	tests := []struct {
		name        string
		initialName string
		setNameTo   string
		want        string
	}{
		{
			name:        "new-pluginName",
			initialName: "",
			setNameTo:   "new-pluginName",
			want:        "new-pluginName",
		},
		{
			name:        "reset to default",
			initialName: "initial",
			setNameTo:   "",
			want:        defaultTableName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert, require := assert.New(t), require.New(t)
			def := NewPluginExecutable("versionid", "os", "arch", []byte("test"))
			require.Equal(defaultTableName, def.TableName())
			s := &PluginExecutable{
				PluginExecutable: &store.PluginExecutable{},
				tableName:        tt.initialName,
			}
			s.SetTableName(tt.setNameTo)
			assert.Equal(tt.want, s.TableName())
		})
	}
}
