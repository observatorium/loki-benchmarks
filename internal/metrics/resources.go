package metrics

import (
	"fmt"

	"github.com/onsi/gomega/gmeasure"
	"github.com/prometheus/common/model"
)

func ContainerCPU(job string, duration model.Duration, annotation gmeasure.Annotation) Measurement {
	return Measurement{
		Name: "Container CPU Usage",
		Query: fmt.Sprintf(
			`sum(avg_over_time(pod:container_cpu_usage:sum{pod=~".*%s.*"}[%s])) * %d`,
			job, duration, CoresToMillicores,
		),
		Unit:       MillicoresUnit,
		Annotation: annotation,
	}
}

func ContainerMemoryWorkingSetBytes(job string, duration model.Duration, annotation gmeasure.Annotation) Measurement {
	return Measurement{
		Name: "Container WorkingSet Memory",
		Query: fmt.Sprintf(
			`sum(avg_over_time(container_memory_working_set_bytes{pod=~".*%s.*", container=""}[%s]) / %d)`,
			job, duration, BytesToGigabytesMultiplier,
		),
		Unit:       GigabytesUnit,
		Annotation: annotation,
	}
}

func PersistentVolumeUsedBytes(job string, duration model.Duration, annotation gmeasure.Annotation) Measurement {
	return Measurement{
		Name: "Persistent Volume Used Bytes",
		Query: fmt.Sprintf(
			`sum(avg_over_time(kubelet_volume_stats_used_bytes{persistentvolumeclaim=~".*%s.*"}[%s]) / %d)`,
			job, duration, BytesToGigabytesMultiplier,
		),
		Unit:       GigabytesUnit,
		Annotation: annotation,
	}
}
