package node

import "github.com/prometheus/client_golang/prometheus"

var (
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
