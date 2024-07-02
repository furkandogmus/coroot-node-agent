package node

import (
	"encoding/xml"
	"fmt"
	"log"
	"strconv"

	"github.com/coroot/coroot-node-agent/node/libvirtSchema"
	"github.com/prometheus/client_golang/prometheus"
	"libvirt.org/go/libvirt"
)

func LibvirtSetup(libvirtUri string, ch chan<- prometheus.Metric) {
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
