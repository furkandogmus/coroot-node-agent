package node

import (
	"encoding/xml"
	"fmt"
	"log"
	"strconv"

	"github.com/coroot/coroot-node-agent/common"
	"github.com/coroot/coroot-node-agent/flags"
	"github.com/coroot/coroot-node-agent/node/libvirtSchema"
	"github.com/coroot/coroot-node-agent/node/metadata"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"
	"libvirt.org/go/libvirt"
)

var (
	procRoot = "/proc"
	infoDesc = prometheus.NewDesc(
		"node_info",
		"Meta information about the node",
		[]string{"hostname", "kernel_version"}, nil,
	)
	cloudInfoDesc = prometheus.NewDesc(
		"node_cloud_info",
		"Meta information about the cloud instance",
		[]string{"provider", "account_id", "instance_id", "instance_type", "instance_life_cycle", "region", "availability_zone", "availability_zone_id", "local_ipv4", "public_ipv4"}, nil,
	)
	uptimeDesc = prometheus.NewDesc(
		"node_uptime_seconds",
		"Uptime of the node in seconds",
		[]string{}, nil,
	)
	cpuUsageDesc = prometheus.NewDesc(
		"node_resources_cpu_usage_seconds_total",
		"The amount of CPU time spent in each mode",
		[]string{"mode"}, nil,
	)
	cpuLogicalCoresDesc = prometheus.NewDesc(
		"node_resources_cpu_logical_cores",
		"The number of logical CPU cores",
		nil, nil,
	)
	memTotalDesc = prometheus.NewDesc(
		"node_resources_memory_total_bytes",
		"The total amount of physical memory",
		nil, nil,
	)
	memFreeDesc = prometheus.NewDesc(
		"node_resources_memory_free_bytes",
		"The amount of unassigned memory",
		nil, nil,
	)
	memAvailableDesc = prometheus.NewDesc(
		"node_resources_memory_available_bytes",
		"The total amount of available memory",
		nil, nil,
	)
	memCacheDesc = prometheus.NewDesc(
		"node_resources_memory_cached_bytes",
		"The amount of memory used as page cache",
		nil, nil,
	)
	diskReadsDesc = prometheus.NewDesc(
		"node_resources_disk_reads_total",
		"The total number of reads completed successfully",
		[]string{"device"}, nil,
	)
	diskWritesDesc = prometheus.NewDesc(
		"node_resources_disk_writes_total",
		"The total number of writes completed successfully",
		[]string{"device"}, nil,
	)
	diskReadBytesDesc = prometheus.NewDesc(
		"node_resources_disk_read_bytes_total",
		"The total number of bytes read from the disk",
		[]string{"device"}, nil,
	)
	diskWrittenBytesDesc = prometheus.NewDesc(
		"node_resources_disk_written_bytes_total",
		"The total number of bytes written to the disk",
		[]string{"device"}, nil,
	)
	diskReadTimeDesc = prometheus.NewDesc(
		"node_resources_disk_read_time_seconds_total",
		"The total number of seconds spent reading",
		[]string{"device"}, nil,
	)
	diskWriteTimeDesc = prometheus.NewDesc(
		"node_resources_disk_write_time_seconds_total",
		"The total number of seconds spent writing",
		[]string{"device"}, nil,
	)
	diskIoTimeDesc = prometheus.NewDesc(
		"node_resources_disk_io_time_seconds_total",
		"The total number of seconds the disk spent doing I/O",
		[]string{"device"}, nil,
	)
	netRxBytesDesc = prometheus.NewDesc(
		"node_net_received_bytes_total",
		"The total number of bytes received",
		[]string{"interface"}, nil,
	)
	netTxBytesDesc = prometheus.NewDesc(
		"node_net_transmitted_bytes_total",
		"The total number of bytes transmitted",
		[]string{"interface"}, nil,
	)
	netRxPacketsDesc = prometheus.NewDesc(
		"node_net_received_packets_total",
		"The total number of packets received",
		[]string{"interface"}, nil,
	)
	netTxPacketsDesc = prometheus.NewDesc(
		"node_net_transmitted_packets_total",
		"The total number of packets transmitted",
		[]string{"interface"}, nil,
	)
	netIfaceUpDesc = prometheus.NewDesc(
		"node_net_interface_up",
		"Status of the interface (0:down, 1:up)",
		[]string{"interface"}, nil,
	)
	ipDesc = prometheus.NewDesc(
		"node_net_interface_ip",
		"IP address assigned to the interface",
		[]string{"interface", "ip"}, nil,
	)

	libvirtUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "", "up"),
		"Whether scraping libvirt's metrics was successful.",
		nil,
		nil)
	libvirtPoolInfoCapacity = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "pool_info", "capacity_bytes"),
		"Pool capacity, in bytes",
		[]string{"pool"},
		nil)
	libvirtPoolInfoAllocation = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "pool_info", "allocation_bytes"),
		"Pool allocation, in bytes",
		[]string{"pool"},
		nil)
	libvirtPoolInfoAvailable = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "pool_info", "available_bytes"),
		"Pool available, in bytes",
		[]string{"pool"},
		nil)
	libvirtVersionsInfoDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "", "versions_info"),
		"Versions of virtualization components",
		[]string{"hypervisor_running", "libvirtd_running", "libvirt_library"},
		nil)
	libvirtDomainInfoMetaDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_info", "meta"),
		"Domain metadata",
		[]string{"domain", "uuid", "instance_name", "flavor", "user_name", "user_uuid", "project_name", "project_uuid", "root_type", "root_uuid"},
		nil)
	libvirtDomainInfoMaxMemBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_info", "maximum_memory_bytes"),
		"Maximum allowed memory of the domain, in bytes.",
		[]string{"domain"},
		nil)
	libvirtDomainInfoMemoryUsageBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_info", "memory_usage_bytes"),
		"Memory usage of the domain, in bytes.",
		[]string{"domain"},
		nil)
	libvirtDomainInfoNrVirtCPUDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_info", "virtual_cpus"),
		"Number of virtual CPUs for the domain.",
		[]string{"domain"},
		nil)
	libvirtDomainInfoCPUTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_info", "cpu_time_seconds_total"),
		"Amount of CPU time used by the domain, in seconds.",
		[]string{"domain"},
		nil)
	libvirtDomainInfoVirDomainState = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_info", "vstate"),
		"Virtual domain state. 0: no state, 1: the domain is running, 2: the domain is blocked on resource,"+
			" 3: the domain is paused by user, 4: the domain is being shut down, 5: the domain is shut off,"+
			"6: the domain is crashed, 7: the domain is suspended by guest power management",
		[]string{"domain"},
		nil)

	libvirtDomainVcpuTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_vcpu", "time_seconds_total"),
		"Amount of CPU time used by the domain's VCPU, in seconds.",
		[]string{"domain", "vcpu"},
		nil)
	libvirtDomainVcpuDelayDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_vcpu", "delay_seconds_total"),
		"Amount of CPU time used by the domain's VCPU, in seconds. "+
			"Vcpu's delay metric. Time the vcpu thread was enqueued by the "+
			"host scheduler, but was waiting in the queue instead of running. "+
			"Exposed to the VM as a steal time.",
		[]string{"domain", "vcpu"},
		nil)
	libvirtDomainVcpuStateDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_vcpu", "state"),
		"VCPU state. 0: offline, 1: running, 2: blocked",
		[]string{"domain", "vcpu"},
		nil)
	libvirtDomainVcpuCPUDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_vcpu", "cpu"),
		"Real CPU number, or one of the values from virVcpuHostCpuState",
		[]string{"domain", "vcpu"},
		nil)
	libvirtDomainVcpuWaitDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_vcpu", "wait_seconds_total"),
		"Vcpu's wait_sum metric. CONFIG_SCHEDSTATS has to be enabled",
		[]string{"domain", "vcpu"},
		nil)

	libvirtDomainMetaBlockDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block", "meta"),
		"Block device metadata info. Device name, source file, serial.",
		[]string{"domain", "target_device", "source_file", "serial", "bus", "disk_type", "driver_type", "cache", "discard"},
		nil)
	libvirtDomainBlockRdBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "read_bytes_total"),
		"Number of bytes read from a block device, in bytes.",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockRdReqDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "read_requests_total"),
		"Number of read requests from a block device.",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockRdTotalTimeSecondsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "read_time_seconds_total"),
		"Total time spent on reads from a block device, in seconds.",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockWrBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "write_bytes_total"),
		"Number of bytes written to a block device, in bytes.",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockWrReqDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "write_requests_total"),
		"Number of write requests to a block device.",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockWrTotalTimesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "write_time_seconds_total"),
		"Total time spent on writes on a block device, in seconds",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockFlushReqDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "flush_requests_total"),
		"Total flush requests from a block device.",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockFlushTotalTimeSecondsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "flush_time_seconds_total"),
		"Total time in seconds spent on cache flushing to a block device",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockAllocationDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "allocation"),
		"Offset of the highest written sector on a block device.",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockCapacityBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "capacity_bytes"),
		"Logical size in bytes of the block device	backing image.",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockPhysicalSizeBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "physicalsize_bytes"),
		"Physical size in bytes of the container of the backing image.",
		[]string{"domain", "target_device"},
		nil)

	// Block IO tune parameters
	// Limits
	libvirtDomainBlockTotalBytesSecDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "limit_total_bytes"),
		"Total throughput limit in bytes per second",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockWriteBytesSecDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "limit_write_bytes"),
		"Write throughput limit in bytes per second",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockReadBytesSecDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "limit_read_bytes"),
		"Read throughput limit in bytes per second",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockTotalIopsSecDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "limit_total_requests"),
		"Total requests per second limit",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockWriteIopsSecDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "limit_write_requests"),
		"Write requests per second limit",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockReadIopsSecDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "limit_read_requests"),
		"Read requests per second limit",
		[]string{"domain", "target_device"},
		nil)
	// Burst limits
	libvirtDomainBlockTotalBytesSecMaxDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "limit_burst_total_bytes"),
		"Total throughput burst limit in bytes per second",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockWriteBytesSecMaxDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "limit_burst_write_bytes"),
		"Write throughput burst limit in bytes per second",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockReadBytesSecMaxDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "limit_burst_read_bytes"),
		"Read throughput burst limit in bytes per second",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockTotalIopsSecMaxDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "limit_burst_total_requests"),
		"Total requests per second burst limit",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockWriteIopsSecMaxDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "limit_burst_write_requests"),
		"Write requests per second burst limit",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockReadIopsSecMaxDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "limit_burst_read_requests"),
		"Read requests per second burst limit",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockTotalBytesSecMaxLengthDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "limit_burst_total_bytes_length_seconds"),
		"Total throughput burst time in seconds",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockWriteBytesSecMaxLengthDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "limit_burst_write_bytes_length_seconds"),
		"Write throughput burst time in seconds",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockReadBytesSecMaxLengthDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "limit_burst_read_bytes_length_seconds"),
		"Read throughput burst time in seconds",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockTotalIopsSecMaxLengthDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "limit_burst_length_total_requests_seconds"),
		"Total requests per second burst time in seconds",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockWriteIopsSecMaxLengthDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "limit_burst_length_write_requests_seconds"),
		"Write requests per second burst time in seconds",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockReadIopsSecMaxLengthDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "limit_burst_length_read_requests_seconds"),
		"Read requests per second burst time in seconds",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainBlockSizeIopsSecDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "size_iops_bytes"),
		"The size of IO operations per second permitted through a block device",
		[]string{"domain", "target_device"},
		nil)

	libvirtDomainMetaInterfacesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface", "meta"),
		"Interfaces metadata. Source bridge, target device, interface uuid",
		[]string{"domain", "source_bridge", "target_device", "virtual_interface"},
		nil)
	libvirtDomainInterfaceRxBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_bytes_total"),
		"Number of bytes received on a network interface, in bytes.",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainInterfaceRxPacketsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_packets_total"),
		"Number of packets received on a network interface.",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainInterfaceRxErrsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_errors_total"),
		"Number of packet receive errors on a network interface.",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainInterfaceRxDropDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_drops_total"),
		"Number of packet receive drops on a network interface.",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainInterfaceTxBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_bytes_total"),
		"Number of bytes transmitted on a network interface, in bytes.",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainInterfaceTxPacketsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_packets_total"),
		"Number of packets transmitted on a network interface.",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainInterfaceTxErrsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_errors_total"),
		"Number of packet transmit errors on a network interface.",
		[]string{"domain", "target_device"},
		nil)
	libvirtDomainInterfaceTxDropDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_drops_total"),
		"Number of packet transmit drops on a network interface.",
		[]string{"domain", "target_device"},
		nil)

	libvirtDomainMemoryStatMajorFaultTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_stats", "major_fault_total"),
		"Page faults occur when a process makes a valid access to virtual memory that is not available. "+
			"When servicing the page fault, if disk IO is required, it is considered a major fault.",
		[]string{"domain"},
		nil)
	libvirtDomainMemoryStatMinorFaultTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_stats", "minor_fault_total"),
		"Page faults occur when a process makes a valid access to virtual memory that is not available. "+
			"When servicing the page not fault, if disk IO is required, it is considered a minor fault.",
		[]string{"domain"},
		nil)
	libvirtDomainMemoryStatUnusedBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_stats", "unused_bytes"),
		"The amount of memory left completely unused by the system. Memory that is available but used for "+
			"reclaimable caches should NOT be reported as free. This value is expressed in bytes.",
		[]string{"domain"},
		nil)
	libvirtDomainMemoryStatAvailableBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_stats", "available_bytes"),
		"The total amount of usable memory as seen by the domain. This value may be less than the amount of "+
			"memory assigned to the domain if a balloon driver is in use or if the guest OS does not initialize all "+
			"assigned pages. This value is expressed in bytes.",
		[]string{"domain"},
		nil)
	libvirtDomainMemoryStatActualBaloonBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_stats", "actual_balloon_bytes"),
		"Current balloon value (in bytes).",
		[]string{"domain"},
		nil)
	libvirtDomainMemoryStatRssBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_stats", "rss_bytes"),
		"Resident Set Size of the process running the domain. This value is in bytes",
		[]string{"domain"},
		nil)
	libvirtDomainMemoryStatUsableBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_stats", "usable_bytes"),
		"How much the balloon can be inflated without pushing the guest system to swap, corresponds "+
			"to 'Available' in /proc/meminfo",
		[]string{"domain"},
		nil)
	libvirtDomainMemoryStatDiskCachesBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_stats", "disk_cache_bytes"),
		"The amount of memory, that can be quickly reclaimed without additional I/O (in bytes)."+
			"Typically these pages are used for caching files from disk.",
		[]string{"domain"},
		nil)
	libvirtDomainMemoryStatUsedPercentDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_stats", "used_percent"),
		"The amount of memory in percent, that used by domain.",
		[]string{"domain"},
		nil)

	errorsMap map[string]struct{}
)

type MemoryStat struct {
	TotalBytes     float64
	FreeBytes      float64
	AvailableBytes float64
	CachedBytes    float64
}

type CpuStat struct {
	TotalUsage   CpuUsage
	LogicalCores int
}

type CpuUsage struct {
	User    float64
	Nice    float64
	System  float64
	Idle    float64
	IoWait  float64
	Irq     float64
	SoftIrq float64
	Steal   float64
}

type Collector struct {
	hostname         string
	kernelVersion    string
	instanceMetadata *metadata.CloudMetadata
}

func NewCollector(hostname, kernelVersion string) *Collector {
	md := metadata.GetInstanceMetadata()
	klog.Infof("instance metadata: %+v", md)
	return &Collector{
		hostname:         hostname,
		kernelVersion:    kernelVersion,
		instanceMetadata: md,
	}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	ch <- gauge(infoDesc, 1, c.hostname, c.kernelVersion)

	libvirtUri := flags.GetString(flags.LibvirtURI)
	conn, err := libvirt.NewConnect(libvirtUri)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	hypervisorVersionNum, err := conn.GetVersion() // virConnectGetVersion, hypervisor running, e.g. QEMU
	if err != nil {
		panic(err)
	}
	hypervisorVersion := fmt.Sprintf("%d.%d.%d", hypervisorVersionNum/1000000%1000, hypervisorVersionNum/1000%1000, hypervisorVersionNum%1000)

	libvirtdVersionNum, err := conn.GetLibVersion() // virConnectGetLibVersion, libvirt daemon running
	if err != nil {
		panic(err)
	}
	libvirtdVersion := fmt.Sprintf("%d.%d.%d", libvirtdVersionNum/1000000%1000, libvirtdVersionNum/1000%1000, libvirtdVersionNum%1000)

	libraryVersionNum, err := libvirt.GetVersion() // virGetVersion, version of libvirt (dynamic) library used by this binary (exporter), not the daemon version
	if err != nil {
		panic(err)
	}
	libraryVersion := fmt.Sprintf("%d.%d.%d", libraryVersionNum/1000000%1000, libraryVersionNum/1000%1000, libraryVersionNum%1000)

	ch <- prometheus.MustNewConstMetric(
		libvirtVersionsInfoDesc,
		prometheus.GaugeValue,
		1.0,
		hypervisorVersion,
		libvirtdVersion,
		libraryVersion)

	stats, err := conn.GetAllDomainStats([]*libvirt.Domain{}, libvirt.DOMAIN_STATS_STATE|libvirt.DOMAIN_STATS_CPU_TOTAL|
		libvirt.DOMAIN_STATS_INTERFACE|libvirt.DOMAIN_STATS_BALLOON|libvirt.DOMAIN_STATS_BLOCK|
		libvirt.DOMAIN_STATS_PERF|libvirt.DOMAIN_STATS_VCPU,
		libvirt.CONNECT_GET_ALL_DOMAINS_STATS_RUNNING|libvirt.CONNECT_GET_ALL_DOMAINS_STATS_SHUTOFF)
	defer func(stats []libvirt.DomainStats) {
		for _, stat := range stats {
			stat.Domain.Free()
		}
	}(stats)
	if err != nil {
		panic(err)
	}
	for _, stat := range stats {
		err = CollectDomain(ch, stat)
		if err != nil {
			log.Printf("Failed to scrape metrics: %s", err)
		}
	}

	// Collect pool info
	pools, err := conn.ListAllStoragePools(libvirt.CONNECT_LIST_STORAGE_POOLS_ACTIVE)
	if err != nil {
		panic(err)
	}
	for _, pool := range pools {
		err = CollectStoragePool(ch, pool)
		pool.Free()
		if err != nil {
			panic(err)
		}
	}

	v, err := uptime(procRoot)
	if err != nil {
		klog.Errorln(err)
	} else {
		ch <- gauge(uptimeDesc, v)
	}

	cpu, err := cpuStat(procRoot)
	if err != nil {
		if !common.IsNotExist(err) {
			klog.Errorln(err)
		}
	} else {
		ch <- counter(cpuUsageDesc, cpu.TotalUsage.User, "user")
		ch <- counter(cpuUsageDesc, cpu.TotalUsage.Nice, "nice")
		ch <- counter(cpuUsageDesc, cpu.TotalUsage.System, "system")
		ch <- counter(cpuUsageDesc, cpu.TotalUsage.Idle, "idle")
		ch <- counter(cpuUsageDesc, cpu.TotalUsage.IoWait, "iowait")
		ch <- counter(cpuUsageDesc, cpu.TotalUsage.Irq, "irq")
		ch <- counter(cpuUsageDesc, cpu.TotalUsage.SoftIrq, "softirq")
		ch <- counter(cpuUsageDesc, cpu.TotalUsage.Steal, "steal")
		ch <- gauge(cpuLogicalCoresDesc, float64(cpu.LogicalCores))
	}

	mem, err := memoryInfo(procRoot)
	if err != nil {
		if !common.IsNotExist(err) {
			klog.Errorln(err)
		}
	} else {
		ch <- gauge(memTotalDesc, mem.TotalBytes)
		ch <- gauge(memFreeDesc, mem.FreeBytes)
		ch <- gauge(memAvailableDesc, mem.AvailableBytes)
		ch <- gauge(memCacheDesc, mem.CachedBytes)
	}

	disks, err := GetDisks()
	if err != nil {
		klog.Errorln("failed to get disk stats:", err)
	} else {
		for _, d := range disks.BlockDevices() {
			ch <- counter(diskReadsDesc, d.ReadOps, d.Name)
			ch <- counter(diskWritesDesc, d.WriteOps, d.Name)
			ch <- counter(diskReadBytesDesc, d.BytesRead, d.Name)
			ch <- counter(diskWrittenBytesDesc, d.BytesWritten, d.Name)
			ch <- counter(diskReadTimeDesc, d.ReadTimeSeconds, d.Name)
			ch <- counter(diskWriteTimeDesc, d.WriteTimeSeconds, d.Name)
			ch <- counter(diskIoTimeDesc, d.IoTimeSeconds, d.Name)
		}
	}

	netdev, err := NetDevices()
	if err != nil {
		klog.Errorln(err)
	} else {
		for _, dev := range netdev {
			ch <- counter(netRxBytesDesc, dev.RxBytes, dev.Name)
			ch <- counter(netTxBytesDesc, dev.TxBytes, dev.Name)
			ch <- counter(netRxPacketsDesc, dev.RxPackets, dev.Name)
			ch <- counter(netTxPacketsDesc, dev.TxPackets, dev.Name)
			ch <- gauge(netIfaceUpDesc, dev.Up, dev.Name)
			for _, p := range dev.IPPrefixes {
				ch <- gauge(ipDesc, 1, dev.Name, p.IP().String())
			}
		}
	}

	im := metadata.CloudMetadata{}
	if c.instanceMetadata != nil {
		im = *c.instanceMetadata
	}
	if f := flags.GetString(flags.Provider); f != "" {
		im.Provider = metadata.CloudProvider(f)
	}
	if f := flags.GetString(flags.Region); f != "" {
		im.Region = f
	}
	if f := flags.GetString(flags.AvailabilityZone); f != "" {
		im.AvailabilityZone = f
	}
	if f := flags.GetString(flags.InstanceType); f != "" {
		im.InstanceType = f
	}
	if f := flags.GetString(flags.InstanceLifeCycle); f != "" {
		im.LifeCycle = f
	}
	ch <- gauge(cloudInfoDesc, 1,
		string(im.Provider), im.AccountId, im.InstanceId, im.InstanceType, im.LifeCycle,
		im.Region, im.AvailabilityZone, im.AvailabilityZoneId, im.LocalIPv4, im.PublicIPv4,
	)
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- infoDesc
	ch <- cloudInfoDesc
	ch <- uptimeDesc
	ch <- cpuUsageDesc
	ch <- cpuLogicalCoresDesc
	ch <- memTotalDesc
	ch <- memFreeDesc
	ch <- memAvailableDesc
	ch <- memCacheDesc
	ch <- diskReadsDesc
	ch <- diskWritesDesc
	ch <- diskReadBytesDesc
	ch <- diskWrittenBytesDesc
	ch <- diskReadTimeDesc
	ch <- diskWriteTimeDesc
	ch <- diskIoTimeDesc
	ch <- netRxBytesDesc
	ch <- netTxBytesDesc
	ch <- netRxPacketsDesc
	ch <- netTxPacketsDesc
	ch <- netIfaceUpDesc
	ch <- ipDesc
}

func CollectDomain(ch chan<- prometheus.Metric, stat libvirt.DomainStats) error {
	domainName, err := stat.Domain.GetName()
	if err != nil {
		return err
	}

	domainUUID, err := stat.Domain.GetUUIDString()
	if err != nil {
		return err
	}

	// Decode XML description of domain to get block device names, etc.
	xmlDesc, err := stat.Domain.GetXMLDesc(0)
	if err != nil {
		return err
	}
	var desc libvirtSchema.Domain
	err = xml.Unmarshal([]byte(xmlDesc), &desc)
	if err != nil {
		return err
	}

	// Report domain info.
	info, err := stat.Domain.GetInfo()
	if err != nil {
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainInfoMetaDesc,
		prometheus.GaugeValue,
		float64(1),
		domainName,
		domainUUID,
		desc.Metadata.NovaInstance.NovaName,
		desc.Metadata.NovaInstance.NovaFlavor.FlavorName,
		desc.Metadata.NovaInstance.NovaOwner.NovaUser.UserName,
		desc.Metadata.NovaInstance.NovaOwner.NovaUser.UserUUID,
		desc.Metadata.NovaInstance.NovaOwner.NovaProject.ProjectName,
		desc.Metadata.NovaInstance.NovaOwner.NovaProject.ProjectUUID,
		desc.Metadata.NovaInstance.NovaRoot.RootType,
		desc.Metadata.NovaInstance.NovaRoot.RootUUID)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainInfoMaxMemBytesDesc,
		prometheus.GaugeValue,
		float64(info.MaxMem)*1024,
		domainName)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainInfoMemoryUsageBytesDesc,
		prometheus.GaugeValue,
		float64(info.Memory)*1024,
		domainName)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainInfoNrVirtCPUDesc,
		prometheus.GaugeValue,
		float64(info.NrVirtCpu),
		domainName)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainInfoCPUTimeDesc,
		prometheus.CounterValue,
		float64(info.CpuTime)/1000/1000/1000, // From nsec to sec
		domainName)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainInfoVirDomainState,
		prometheus.GaugeValue,
		float64(info.State),
		domainName)

	domainStatsVcpu, err := stat.Domain.GetVcpus()
	if err != nil {
		lverr, ok := err.(libvirt.Error)
		if !ok || lverr.Code != libvirt.ERR_OPERATION_INVALID {
			return err
		}
	} else {
		for _, vcpu := range domainStatsVcpu {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainVcpuStateDesc,
				prometheus.GaugeValue,
				float64(vcpu.State),
				domainName,
				strconv.FormatInt(int64(vcpu.Number), 10))

			ch <- prometheus.MustNewConstMetric(
				libvirtDomainVcpuTimeDesc,
				prometheus.CounterValue,
				float64(vcpu.CpuTime)/1000/1000/1000, // From nsec to sec
				domainName,
				strconv.FormatInt(int64(vcpu.Number), 10))

			ch <- prometheus.MustNewConstMetric(
				libvirtDomainVcpuCPUDesc,
				prometheus.GaugeValue,
				float64(vcpu.Cpu),
				domainName,
				strconv.FormatInt(int64(vcpu.Number), 10))
		}
		/* There's no Wait in GetVcpus()
		 * But there's no cpu number in libvirt.DomainStats
		 * Time and State are present in both structs
		 * So, let's take Wait here
		 */
		for cpuNum, vcpu := range stat.Vcpu {
			if vcpu.WaitSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainVcpuWaitDesc,
					prometheus.CounterValue,
					float64(vcpu.Wait)/1000/1000/1000,
					domainName,
					strconv.FormatInt(int64(cpuNum), 10))
			}
			if vcpu.DelaySet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainVcpuDelayDesc,
					prometheus.CounterValue,
					float64(vcpu.Delay)/1e9,
					domainName,
					strconv.FormatInt(int64(cpuNum), 10))
			}
		}
	}

	// Report block device statistics.
	for _, disk := range stat.Block {
		var DiskSource string
		var Device *libvirtSchema.Disk
		// Ugly hack to avoid getting metrics from cdrom block device
		// TODO: somehow check the disk 'device' field for 'cdrom' string
		if disk.Name == "hdc" || disk.Name == "hda" {
			continue
		}
		/*  "block.<num>.path" - string describing the source of block device <num>,
		    if it is a file or block device (omitted for network
		    sources and drives with no media inserted). For network device (i.e. rbd) take from xml. */
		for _, dev := range desc.Devices.Disks {
			if dev.Target.Device == disk.Name {
				if disk.PathSet {
					DiskSource = disk.Path

				} else {
					DiskSource = dev.Source.Name
				}
				Device = &dev
				break
			}
		}

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainMetaBlockDesc,
			prometheus.GaugeValue,
			float64(1),
			domainName,
			disk.Name,
			DiskSource,
			Device.Serial,
			Device.Target.Bus,
			Device.DiskType,
			Device.Driver.Type,
			Device.Driver.Cache,
			Device.Driver.Discard,
		)

		// https://libvirt.org/html/libvirt-libvirt-domain.html#virConnectGetAllDomainStats
		if disk.RdBytesSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockRdBytesDesc,
				prometheus.CounterValue,
				float64(disk.RdBytes),
				domainName,
				disk.Name)
		}
		if disk.RdReqsSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockRdReqDesc,
				prometheus.CounterValue,
				float64(disk.RdReqs),
				domainName,
				disk.Name)
		}
		if disk.RdTimesSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockRdTotalTimeSecondsDesc,
				prometheus.CounterValue,
				float64(disk.RdTimes)/1e9,
				domainName,
				disk.Name)
		}
		if disk.WrBytesSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockWrBytesDesc,
				prometheus.CounterValue,
				float64(disk.WrBytes),
				domainName,
				disk.Name)
		}
		if disk.WrReqsSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockWrReqDesc,
				prometheus.CounterValue,
				float64(disk.WrReqs),
				domainName,
				disk.Name)
		}
		if disk.WrTimesSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockWrTotalTimesDesc,
				prometheus.CounterValue,
				float64(disk.WrTimes)/1e9,
				domainName,
				disk.Name)
		}
		if disk.FlReqsSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockFlushReqDesc,
				prometheus.CounterValue,
				float64(disk.FlReqs),
				domainName,
				disk.Name)
		}
		if disk.FlTimesSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockFlushTotalTimeSecondsDesc,
				prometheus.CounterValue,
				float64(disk.FlTimes)/1e9,
				domainName,
				disk.Name)
		}
		if disk.AllocationSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockAllocationDesc,
				prometheus.GaugeValue,
				float64(disk.Allocation),
				domainName,
				disk.Name)
		}
		if disk.CapacitySet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockCapacityBytesDesc,
				prometheus.GaugeValue,
				float64(disk.Capacity),
				domainName,
				disk.Name)
		}
		if disk.PhysicalSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainBlockPhysicalSizeBytesDesc,
				prometheus.GaugeValue,
				float64(disk.Physical),
				domainName,
				disk.Name)
		}

		blockIOTuneParams, err := stat.Domain.GetBlockIoTune(disk.Name, 0)
		if err != nil {
			lverr, ok := err.(libvirt.Error)
			if !ok {
				switch lverr.Code {
				case libvirt.ERR_OPERATION_INVALID:
					// This should be one-shot error
					log.Printf("Invalid operation GetBlockIoTune: %s", err.Error())
				case libvirt.ERR_OPERATION_UNSUPPORTED:
					WriteErrorOnce("Unsupported operation GetBlockIoTune: "+err.Error(), "blkiotune_unsupported")
				default:
					return err
				}
			}
		} else {
			if blockIOTuneParams.TotalBytesSecSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockTotalBytesSecDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.TotalBytesSec),
					domainName,
					disk.Name)
			}
			if blockIOTuneParams.ReadBytesSecSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockReadBytesSecDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.ReadBytesSec),
					domainName,
					disk.Name)
			}
			if blockIOTuneParams.WriteBytesSecSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockWriteBytesSecDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.WriteBytesSec),
					domainName,
					disk.Name)
			}
			if blockIOTuneParams.TotalIopsSecSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockTotalIopsSecDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.TotalIopsSec),
					domainName,
					disk.Name)
			}
			if blockIOTuneParams.ReadIopsSecSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockReadIopsSecDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.ReadIopsSec),
					domainName,
					disk.Name)
			}
			if blockIOTuneParams.WriteIopsSecSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockWriteIopsSecDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.WriteIopsSec),
					domainName,
					disk.Name)
			}
			if blockIOTuneParams.TotalBytesSecMaxSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockTotalBytesSecMaxDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.TotalBytesSecMax),
					domainName,
					disk.Name)
			}
			if blockIOTuneParams.ReadBytesSecMaxSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockReadBytesSecMaxDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.ReadBytesSecMax),
					domainName,
					disk.Name)
			}
			if blockIOTuneParams.WriteBytesSecMaxSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockWriteBytesSecMaxDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.WriteBytesSecMax),
					domainName,
					disk.Name)
			}
			if blockIOTuneParams.TotalIopsSecMaxSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockTotalIopsSecMaxDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.TotalIopsSecMax),
					domainName,
					disk.Name)
			}
			if blockIOTuneParams.ReadIopsSecMaxSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockReadIopsSecMaxDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.ReadIopsSecMax),
					domainName,
					disk.Name)
			}
			if blockIOTuneParams.WriteIopsSecMaxSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockWriteIopsSecMaxDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.WriteIopsSecMax),
					domainName,
					disk.Name)
			}
			if blockIOTuneParams.TotalBytesSecMaxLengthSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockTotalBytesSecMaxLengthDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.TotalBytesSecMaxLength),
					domainName,
					disk.Name)
			}
			if blockIOTuneParams.ReadBytesSecMaxLengthSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockReadBytesSecMaxLengthDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.ReadBytesSecMaxLength),
					domainName,
					disk.Name)
			}
			if blockIOTuneParams.WriteBytesSecMaxLengthSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockWriteBytesSecMaxLengthDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.WriteBytesSecMaxLength),
					domainName,
					disk.Name)
			}
			if blockIOTuneParams.TotalIopsSecMaxLengthSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockTotalIopsSecMaxLengthDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.TotalIopsSecMaxLength),
					domainName,
					disk.Name)
			}
			if blockIOTuneParams.ReadIopsSecMaxLengthSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockReadIopsSecMaxLengthDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.ReadIopsSecMaxLength),
					domainName,
					disk.Name)
			}
			if blockIOTuneParams.WriteIopsSecMaxLengthSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockWriteIopsSecMaxLengthDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.WriteIopsSecMaxLength),
					domainName,
					disk.Name)
			}
			if blockIOTuneParams.SizeIopsSecSet {
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainBlockSizeIopsSecDesc,
					prometheus.GaugeValue,
					float64(blockIOTuneParams.SizeIopsSec),
					domainName,
					disk.Name)
			}
		}
	}

	// Report network interface statistics.
	for _, iface := range stat.Net {
		var SourceBridge string
		var VirtualInterface string
		// Additional info for ovs network
		for _, net := range desc.Devices.Interfaces {
			if net.Target.Device == iface.Name {
				SourceBridge = net.Source.Bridge
				VirtualInterface = net.Virtualport.Parameters.InterfaceID
				break
			}
		}
		if SourceBridge != "" || VirtualInterface != "" {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainMetaInterfacesDesc,
				prometheus.GaugeValue,
				float64(1),
				domainName,
				SourceBridge,
				iface.Name,
				VirtualInterface)
		}
		if iface.RxBytesSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainInterfaceRxBytesDesc,
				prometheus.CounterValue,
				float64(iface.RxBytes),
				domainName,
				iface.Name)
		}
		if iface.RxPktsSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainInterfaceRxPacketsDesc,
				prometheus.CounterValue,
				float64(iface.RxPkts),
				domainName,
				iface.Name)
		}
		if iface.RxErrsSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainInterfaceRxErrsDesc,
				prometheus.CounterValue,
				float64(iface.RxErrs),
				domainName,
				iface.Name)
		}
		if iface.RxDropSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainInterfaceRxDropDesc,
				prometheus.CounterValue,
				float64(iface.RxDrop),
				domainName,
				iface.Name)
		}
		if iface.TxBytesSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainInterfaceTxBytesDesc,
				prometheus.CounterValue,
				float64(iface.TxBytes),
				domainName,
				iface.Name)
		}
		if iface.TxPktsSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainInterfaceTxPacketsDesc,
				prometheus.CounterValue,
				float64(iface.TxPkts),
				domainName,
				iface.Name)
		}
		if iface.TxErrsSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainInterfaceTxErrsDesc,
				prometheus.CounterValue,
				float64(iface.TxErrs),
				domainName,
				iface.Name)
		}
		if iface.TxDropSet {
			ch <- prometheus.MustNewConstMetric(
				libvirtDomainInterfaceTxDropDesc,
				prometheus.CounterValue,
				float64(iface.TxDrop),
				domainName,
				iface.Name)
		}
	}

	// Collect Memory Stats
	memorystat, err := stat.Domain.MemoryStats(11, 0)
	var MemoryStats libvirtSchema.VirDomainMemoryStats
	var usedPercent float64
	if err == nil {
		MemoryStats = memoryStatCollect(&memorystat)
		if MemoryStats.Usable != 0 && MemoryStats.Available != 0 {
			usedPercent = (float64(MemoryStats.Available) - float64(MemoryStats.Usable)) / (float64(MemoryStats.Available) / float64(100))
		}

	}
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainMemoryStatMajorFaultTotalDesc,
		prometheus.CounterValue,
		float64(MemoryStats.MajorFault),
		domainName)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainMemoryStatMinorFaultTotalDesc,
		prometheus.CounterValue,
		float64(MemoryStats.MinorFault),
		domainName)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainMemoryStatUnusedBytesDesc,
		prometheus.GaugeValue,
		float64(MemoryStats.Unused)*1024,
		domainName)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainMemoryStatAvailableBytesDesc,
		prometheus.GaugeValue,
		float64(MemoryStats.Available)*1024,
		domainName)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainMemoryStatActualBaloonBytesDesc,
		prometheus.GaugeValue,
		float64(MemoryStats.ActualBalloon)*1024,
		domainName)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainMemoryStatRssBytesDesc,
		prometheus.GaugeValue,
		float64(MemoryStats.Rss)*1024,
		domainName)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainMemoryStatUsableBytesDesc,
		prometheus.GaugeValue,
		float64(MemoryStats.Usable)*1024,
		domainName)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainMemoryStatDiskCachesBytesDesc,
		prometheus.GaugeValue,
		float64(MemoryStats.DiskCaches)*1024,
		domainName)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainMemoryStatUsedPercentDesc,
		prometheus.GaugeValue,
		float64(usedPercent),
		domainName)

	return nil
}

func WriteErrorOnce(err string, name string) {
	if _, ok := errorsMap[name]; !ok {
		log.Printf("%s", err)
		errorsMap[name] = struct{}{}
	}
}

func memoryStatCollect(memorystat *[]libvirt.DomainMemoryStat) libvirtSchema.VirDomainMemoryStats {
	var MemoryStats libvirtSchema.VirDomainMemoryStats
	for _, domainmemorystat := range *memorystat {
		switch tag := domainmemorystat.Tag; tag {
		case 2:
			MemoryStats.MajorFault = domainmemorystat.Val
		case 3:
			MemoryStats.MinorFault = domainmemorystat.Val
		case 4:
			MemoryStats.Unused = domainmemorystat.Val
		case 5:
			MemoryStats.Available = domainmemorystat.Val
		case 6:
			MemoryStats.ActualBalloon = domainmemorystat.Val
		case 7:
			MemoryStats.Rss = domainmemorystat.Val
		case 8:
			MemoryStats.Usable = domainmemorystat.Val
		case 10:
			MemoryStats.DiskCaches = domainmemorystat.Val
		}
	}
	return MemoryStats
}

func CollectStoragePool(ch chan<- prometheus.Metric, pool libvirt.StoragePool) error {
	// Refresh pool
	err := pool.Refresh(0)
	if err != nil {
		return err
	}
	pool_name, err := pool.GetName()
	if err != nil {
		return err
	}
	pool_info, err := pool.GetInfo()
	if err != nil {
		return err
	}
	// Send metrics to channel
	ch <- prometheus.MustNewConstMetric(
		libvirtPoolInfoCapacity,
		prometheus.GaugeValue,
		float64(pool_info.Capacity),
		pool_name)
	ch <- prometheus.MustNewConstMetric(
		libvirtPoolInfoAllocation,
		prometheus.GaugeValue,
		float64(pool_info.Allocation),
		pool_name)
	ch <- prometheus.MustNewConstMetric(
		libvirtPoolInfoAvailable,
		prometheus.GaugeValue,
		float64(pool_info.Available),
		pool_name)
	return nil
}

func counter(desc *prometheus.Desc, value float64, labelValues ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(desc, prometheus.CounterValue, value, labelValues...)
}

func gauge(desc *prometheus.Desc, value float64, labelValues ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value, labelValues...)
}
