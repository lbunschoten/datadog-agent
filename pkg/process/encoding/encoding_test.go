package encoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	model "github.com/DataDog/agent-payload/process"
	"github.com/DataDog/datadog-agent/pkg/process/procutil"
)

func TestSerialization(t *testing.T) {
	origin := map[int32]*procutil.StatsWithPerm{
		1: {
			OpenFdCount: 1,
			IOStat: &procutil.IOCountersStat{
				ReadCount:  1,
				WriteCount: 2,
				ReadBytes:  3,
				WriteBytes: 4,
			},
		},
	}
	exp := &model.ProcStatsWithPermByPID{
		StatsByPID: map[int32]*model.ProcStatsWithPerm{
			1: {
				OpenFDCount: 1,
				ReadCount:   1,
				WriteCount:  2,
				ReadBytes:   3,
				WriteBytes:  4,
			},
		},
	}

	origin2 := map[int32]*procutil.StatsWithPerm{
		1: {
			OpenFdCount: 2,
			IOStat: &procutil.IOCountersStat{
				ReadCount:  4,
				WriteCount: 2,
				ReadBytes:  5,
				WriteBytes: 8,
			},
		},
	}
	exp2 := &model.ProcStatsWithPermByPID{
		StatsByPID: map[int32]*model.ProcStatsWithPerm{
			1: {
				OpenFDCount: 2,
				ReadCount:   4,
				WriteCount:  2,
				ReadBytes:   5,
				WriteBytes:  8,
			},
		},
	}

	t.Run("requesting application/json serialization", func(t *testing.T) {
		marshaler := GetMarshaler("application/json")
		assert.Equal(t, "application/json", marshaler.ContentType())
		blob, err := marshaler.Marshal(origin)
		require.NoError(t, err)

		unmarshaler := GetUnmarshaler("application/json")
		result, err := unmarshaler.Unmarshal(blob)
		require.NoError(t, err)
		assert.Equal(t, exp, result)
	})

	t.Run("requesting empty marshaler name serialization", func(t *testing.T) {
		marshaler := GetMarshaler("")
		// in case we request empty serialization type, default to application/json
		assert.Equal(t, "application/json", marshaler.ContentType())
		blob, err := marshaler.Marshal(origin2)
		require.NoError(t, err)

		unmarshaler := GetUnmarshaler("application/json")
		result, err := unmarshaler.Unmarshal(blob)
		require.NoError(t, err)
		assert.Equal(t, exp2, result)
	})

	t.Run("requesting application/protobuf serialization", func(t *testing.T) {
		marshaler := GetMarshaler("application/protobuf")
		assert.Equal(t, "application/protobuf", marshaler.ContentType())

		blob, err := marshaler.Marshal(origin)
		require.NoError(t, err)

		unmarshaler := GetUnmarshaler("application/protobuf")
		result, err := unmarshaler.Unmarshal(blob)
		require.NoError(t, err)
		assert.Equal(t, exp, result)
	})

	t.Run("protobuf serializing empty input", func(t *testing.T) {
		marshaler := GetMarshaler("application/protobuf")
		assert.Equal(t, "application/protobuf", marshaler.ContentType())

		var empty map[int32]*procutil.StatsWithPerm
		blob, err := marshaler.Marshal(empty)
		require.NoError(t, err)

		unmarshaler := GetUnmarshaler("application/protobuf")
		result, err := unmarshaler.Unmarshal(blob)
		require.NoError(t, err)
		assert.EqualValues(t, &model.ProcStatsWithPermByPID{}, result)
	})

	t.Run("json serializing empty input", func(t *testing.T) {
		marshaler := GetMarshaler("application/json")
		assert.Equal(t, "application/json", marshaler.ContentType())

		var empty map[int32]*procutil.StatsWithPerm
		blob, err := marshaler.Marshal(empty)
		require.NoError(t, err)

		unmarshaler := GetUnmarshaler("application/json")
		result, err := unmarshaler.Unmarshal(blob)
		require.NoError(t, err)
		assert.EqualValues(t, &model.ProcStatsWithPermByPID{StatsByPID: map[int32]*model.ProcStatsWithPerm{}}, result)
	})
}