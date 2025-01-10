package cpuset_test

import (
	"testing"

	"github.com/buildbuddy-io/buildbuddy/enterprise/server/util/cpuset"
	"github.com/buildbuddy-io/buildbuddy/server/util/testing/flags"
	"github.com/buildbuddy-io/buildbuddy/server/util/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCPUSetAcquireRelease(t *testing.T) {
	flags.Set(t, "executor.cpu_leaser.enable", true)
	flags.Set(t, "executor.cpu_leaser.overhead", 0)
	flags.Set(t, "executor.cpu_leaser.min_overhead", 0)
	flags.Set(t, "executor.cpu_leaser.cpuset", "0-3")

	cs, err := cpuset.NewLeaser()
	require.NoError(t, err)
	task1 := uuid.New()
	cpus1, cancel1 := cs.Acquire(3000, task1)
	defer cancel1()
	assert.Equal(t, 3, len(cpus1))

	task2 := uuid.New()
	cpus2, cancel2 := cs.Acquire(1000, task2)
	defer cancel2()
	require.Equal(t, 1, len(cpus2))

	assert.NotContains(t, cpus1, cpus2[0])
}

func TestEvenDistributionUnderLoad(t *testing.T) {
	flags.Set(t, "executor.cpu_leaser.enable", true)
	flags.Set(t, "executor.cpu_leaser.overhead", 0)
	flags.Set(t, "executor.cpu_leaser.min_overhead", 0)
	flags.Set(t, "executor.cpu_leaser.cpuset", "0-3")

	cs, err := cpuset.NewLeaser()
	require.NoError(t, err)

	counts := make(map[int]int, 4)
	for i := 0; i < 100; i++ {
		task := uuid.New()
		cpus, cancel := cs.Acquire(1000, task)
		for _, cpu := range cpus {
			counts[cpu] += 1
		}
		// Cancel when the test is over.
		defer cancel()
	}
	for _, count := range counts {
		assert.Equal(t, 25, count)
	}
}

func TestCPUSetOverhead(t *testing.T) {
	flags.Set(t, "executor.cpu_leaser.enable", true)
	flags.Set(t, "executor.cpu_leaser.overhead", .20)
	flags.Set(t, "executor.cpu_leaser.min_overhead", 2)
	flags.Set(t, "executor.cpu_leaser.cpuset", "0-47")

	cs, err := cpuset.NewLeaser()
	require.NoError(t, err)
	task1 := uuid.New()
	cpus1, cancel1 := cs.Acquire(1100, task1)
	defer cancel1()
	assert.Equal(t, 4, len(cpus1))

	task2 := uuid.New()
	cpus2, cancel2 := cs.Acquire(1800, task2)
	defer cancel2()
	require.Equal(t, 4, len(cpus2))

	assert.NotContains(t, cpus1, cpus2[0])

	task3 := uuid.New()
	cpus3, cancel3 := cs.Acquire(20_000, task3)
	defer cancel3()
	require.Equal(t, 24, len(cpus3))
}

func TestCPUSetDisabled(t *testing.T) {
	flags.Set(t, "executor.cpu_leaser.enable", false)
	flags.Set(t, "executor.cpu_leaser.cpuset", "0-47")

	cs, err := cpuset.NewLeaser()
	require.NoError(t, err)
	task1 := uuid.New()
	cpus1, cancel1 := cs.Acquire(1100, task1)
	defer cancel1()
	assert.Equal(t, 48, len(cpus1))
}

func TestCPUSetDisabledManualCPUSet(t *testing.T) {
	flags.Set(t, "executor.cpu_leaser.enable", false)
	flags.Set(t, "executor.cpu_leaser.cpuset", "0-1,3")

	cs, err := cpuset.NewLeaser()
	require.NoError(t, err)
	task1 := uuid.New()
	cpus1, cancel1 := cs.Acquire(1100, task1)
	defer cancel1()
	assert.Equal(t, []int{0, 1, 3}, cpus1)
}

func TestNumaNodeFairness(t *testing.T) {
	flags.Set(t, "executor.cpu_leaser.enable", true)
	flags.Set(t, "executor.cpu_leaser.overhead", 0)
	flags.Set(t, "executor.cpu_leaser.min_overhead", 0)
	flags.Set(t, "executor.cpu_leaser.cpuset", "0:0-1,1:4-7,0:2-3")

	var allCPUs []int

	cs, err := cpuset.NewLeaser()
	require.NoError(t, err)
	task1 := uuid.New()
	cpus1, cancel1 := cs.Acquire(4000, task1)
	defer cancel1()
	allCPUs = append(allCPUs, cpus1...)

	task2 := uuid.New()
	cpus2, cancel2 := cs.Acquire(4000, task2)
	defer cancel2()
	allCPUs = append(allCPUs, cpus2...)

	assert.ElementsMatch(t, []int{0, 1, 2, 3, 4, 5, 6, 7}, allCPUs)

	task3 := uuid.New()
	cpus3, cancel3 := cs.Acquire(8000, task3)
	defer cancel3()
	assert.Equal(t, len(cpus3), 4)
}

func TestNumaNodesAreNotSplit(t *testing.T) {
	flags.Set(t, "executor.cpu_leaser.enable", true)
	flags.Set(t, "executor.cpu_leaser.overhead", 0)
	flags.Set(t, "executor.cpu_leaser.min_overhead", 0)
	flags.Set(t, "executor.cpu_leaser.cpuset", "0:0-3,1:4-7")

	cs, err := cpuset.NewLeaser()
	require.NoError(t, err)
	task1 := uuid.New()
	cpus1, cancel1 := cs.Acquire(4000, task1)
	defer cancel1()
	assert.Equal(t, 4, len(cpus1))
}

func TestNonContiguousNumaNodes(t *testing.T) {
	flags.Set(t, "executor.cpu_leaser.enable", true)
	flags.Set(t, "executor.cpu_leaser.overhead", 0)
	flags.Set(t, "executor.cpu_leaser.min_overhead", 0)
	flags.Set(t, "executor.cpu_leaser.cpuset", "3:0-3,9:4-7")

	cs, err := cpuset.NewLeaser()
	require.NoError(t, err)

	task1 := uuid.New()
	cpus1, cancel1 := cs.Acquire(4000, task1)
	defer cancel1()

	task2 := uuid.New()
	cpus2, cancel2 := cs.Acquire(4000, task2)
	defer cancel2()

	assert.NotElementsMatch(t, cpus1, cpus2)
}
