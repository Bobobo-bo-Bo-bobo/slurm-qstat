package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func printPartitionStatus(p []partitionInfo, brief bool) {
	var data [][]string
	var idleSum uint64
	var allocatedSum uint64
	var otherSum uint64
	var totalSum uint64

	for _, value := range p {
		idleSum += value.CoresIdle
		allocatedSum += value.CoresAllocated
		otherSum += value.CoresOther
		totalSum += value.CoresTotal

		if brief {
			data = append(data, []string{
				value.Name,
				strconv.FormatUint(value.CoresIdle, 10),
				strconv.FormatUint(value.CoresAllocated, 10),
				strconv.FormatUint(value.CoresOther, 10),
				strconv.FormatUint(value.CoresTotal, 10),
			})
		} else {
			var ipct float64
			var apct float64
			var opct float64

			if value.CoresTotal != 0 {
				ipct = 100.0 * float64(value.CoresIdle) / float64(value.CoresTotal)
				apct = 100.0 * float64(value.CoresAllocated) / float64(value.CoresTotal)
				opct = 100.0 * float64(value.CoresOther) / float64(value.CoresTotal)
			}
			data = append(data, []string{
				value.Name,
				strconv.FormatUint(value.CoresIdle, 10),
				strconv.FormatUint(value.CoresAllocated, 10),
				strconv.FormatUint(value.CoresOther, 10),
				strconv.FormatUint(value.CoresTotal, 10),
				strconv.FormatFloat(ipct, 'f', 3, 64),
				strconv.FormatFloat(apct, 'f', 3, 64),
				strconv.FormatFloat(opct, 'f', 3, 64),
				strconv.FormatFloat(100.0, 'f', 3, 64),
			})
		}
	}

	table := tablewriter.NewWriter(os.Stdout)

	if brief {
		table.SetHeader([]string{"Partition", "Idle", "Allocated", "Other", "Total"})
	} else {
		table.SetHeader([]string{"Partition", "Idle", "Allocated", "Other", "Total", "Idle%", "Allocated%", "Other%", "Total%"})
	}

	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)

	if !brief {
		table.SetFooter([]string{
			"Sum",
			strconv.FormatUint(idleSum, 10),
			strconv.FormatUint(allocatedSum, 10),
			strconv.FormatUint(otherSum, 10),
			strconv.FormatUint(totalSum, 10),

			strconv.FormatFloat(float64(idleSum)/float64(totalSum)*100.0, 'f', 3, 64),
			strconv.FormatFloat(float64(allocatedSum)/float64(totalSum)*100.0, 'f', 3, 64),
			strconv.FormatFloat(float64(otherSum)/float64(totalSum)*100.0, 'f', 3, 64),
			strconv.FormatFloat(100.0, 'f', 3, 64),
		})
		table.SetFooterAlignment(tablewriter.ALIGN_RIGHT)
	}

	table.AppendBulk(data)
	table.Render()

}

func printJobStatus(j []jobData, brief bool) {
	var reUser = regexp.MustCompile(`\(\d+\)`)
	var data [][]string
	var runCount uint64
	var pendCount uint64
	var otherCount uint64
	var totalCount uint64
	var failCount uint64
	var preeemptCount uint64
	var stopCount uint64
	var suspendCount uint64

	for _, jData := range j {
		var host string
		var startTime string
		var pendingReason string

		job, found := jData["JobId"]
		if !found {
			log.Panic("BUG: No job ID found for job\n")
		}

		user, found := jData["UserId"]
		if !found {
			log.Panicf("BUG: No user found for job %s\n", job)
		}

		user = reUser.ReplaceAllString(user, "")

		state, found := jData["JobState"]
		if !found {
			log.Panicf("BUG: No JobState found for job %s\n", job)
		}

		switch state {
		case "FAILED":
			failCount++
		case "PENDING":
			pendCount++
		case "PREEMPTED":
			preeemptCount++
		case "STOPPED":
			stopCount++
		case "SUSPENDED":
			suspendCount++
		case "RUNNING":
			runCount++
		default:
			otherCount++
		}
		totalCount++

		partition, found := jData["Partition"]
		if !found {
			log.Panicf("BUG: No partition found for job %s\n", job)
		}

		tres := jData["TRES"]

		_numCpus, found := jData["NumCPUs"]
		if !found {
			log.Panicf("BUG: NumCPUs not found for job %s\n", job)
		}
		numCpus, err := strconv.ParseUint(_numCpus, 10, 64)
		if err != nil {
			log.Panicf("BUG: Can't convert NumCpus to an integer for job %s: %s\n", job, err)
		}

		name, found := jData["JobName"]
		if !found {
			log.Panicf("BUG: JobName not set for job %s\n", job)
		}

		nodes, found := jData["NodeList"]
		if !found {
			log.Panicf("BUG: NodeList not set for job %s\n", job)
		}
		if nodes == "(null}" {
			nodes = ""
		}

		licenses := jData["Licenses"]
		if licenses == "(null)" {
			licenses = ""
		}

		gres := jData["Gres"]
		if gres == "(null)" {
			gres = ""
		}

		tres = jData["TRES"]
		if tres == "(null}" {
			tres = ""
		}

		if state == "PENDING" {
			// Jobs can also be submitted, requesting a number of Nodes instead of CPUs
			// Therefore we will check TRES first
			tresCpus, err := getCpusFromTresString(tres)
			if err != nil {
				log.Panicf("BUG: Can't get number of CPUs from TRES as integer for job %s: %s\n", job, err)
			}

			if tresCpus != 0 {
				numCpus = tresCpus
			}

			// PENDING jobs never scheduled at all don't have BatchHost set (for obvious reasons)
			// Rescheduled and now PENDING jobs do have a BatchHost
			host, found = jData["BatchHost"]
			if !found {
				host = "<not_scheduled_yet>"
			}

			// The same applies for StartTime
			startTime, found = jData["StartTime"]
			if !found {
				startTime = "<not_scheduled_yet>"
			}

			// Obviously, PENDING jobs _always_ have a Reason
			pendingReason, found = jData["Reason"]
			if !found {
				log.Panicf("BUG: No Reason for pending job %s\n", job)
			}

			nodes = "<not_scheduled_yet>"

		} else {
			host, found = jData["BatchHost"]
			if state == "RUNNING" {
				if !found {
					log.Panicf("BUG: No BatchHost set for job %s\n", job)
				}
			}

			startTime, found = jData["StartTime"]
			if state == "RUNNING" {
				if !found {
					log.Panicf("BUG: No StartTime set for job %s\n", job)
				}
			}
		}

		if brief {
			data = append(data, []string{
				job,
				partition,
				user,
				state,
				nodes,
				strconv.FormatUint(numCpus, 10),
				startTime,
				name,
			})
		} else {
			data = append(data, []string{
				job,
				partition,
				user,
				state,
				pendingReason,
				host,
				nodes,
				strconv.FormatUint(numCpus, 10),
				licenses,
				gres,
				tres,
				startTime,
				name,
			})
		}
	}

	table := tablewriter.NewWriter(os.Stdout)

	if brief {
		table.SetHeader([]string{"JobID", "Partition", "User", "State", "Nodes", "CPUs", "Starttime", "Name"})
	} else {
		table.SetHeader([]string{"JobID", "Partition", "User", "State", "Reason", "Batchhost", "Nodes", "CPUs", "Licenses", "GRES", "TRES", "Starttime", "Name"})
	}

	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)

	if !brief {
		table.SetFooter([]string{
			"Sum",
			"",
			"",
			"",
			"",
			fmt.Sprintf("Failed: %d", failCount),
			fmt.Sprintf("Pending: %d", pendCount),
			fmt.Sprintf("Preempted: %d", preeemptCount),
			fmt.Sprintf("Stoped: %d", stopCount),
			fmt.Sprintf("Suspended: %d", suspendCount),
			fmt.Sprintf("Running: %d", runCount),
			fmt.Sprintf("Other: %d", otherCount),
			fmt.Sprintf("Total: %d", totalCount),
		})
		table.SetFooterAlignment(tablewriter.ALIGN_LEFT)
	}

	table.AppendBulk(data)
	table.Render()
}

func printNodeStatus(n []nodeData, brief bool) {
	var data [][]string
	var totalCount uint64
	var allocCount uint64
	var drainingCount uint64
	var idleCount uint64
	var drainedCount uint64
	var mixedCount uint64
	var downCount uint64
	var otherCount uint64
	var reservedCount uint64

	for _, ndata := range n {
		partitions, found := ndata["Partitions"]
		if !found {
			// Note: Although seldom configured, it is a valid configuration to define a node in SLURM without assiging it to a partition
			partitions = ""
		}

		nname, found := ndata["NodeName"]
		if !found {
			log.Panicf("BUG: No NodeName found for node %+v\n", ndata)
		}
		node := nname

		state, found := ndata["State"]
		if !found {
			log.Panicf("BUG: No State for node %s\n", node)
		}

		if state == "ALLOCATED" {
			allocCount++
		} else if state == "ALLOCATED+DRAIN" {
			drainingCount++
		} else if state == "IDLE" {
			idleCount++
		} else if state == "IDLE+DRAIN" {
			drainedCount++
		} else if state == "MIXED" {
			mixedCount++
		} else if state == "MIXED+DRAIN" {
			drainingCount++
		} else if strings.Contains(state, "DOWN") {
			downCount++
		} else if state == "RESERVED" {
			reservedCount++
		} else {
			otherCount++
		}

		totalCount++

		version := ndata["Version"]

		cfgTres, found := ndata["CfgTRES"]
		if !found {
			log.Panicf("BUG: No CfgTRES for node %s\n", node)
		}

		allocTres, found := ndata["AllocTRES"]
		if !found {
			log.Panicf("BUG: No AllocTRES for node %s\n", node)
		}

		sockets, found := ndata["Sockets"]
		if !found {
			log.Panicf("BUG: No Sockets for node %s\n", node)
		}

		boards, found := ndata["Boards"]
		if !found {
			log.Panicf("BUG: No Boards for node %s\n", node)
		}

		tpc, found := ndata["ThreadsPerCore"]
		if !found {
			log.Panicf("BUG: No ThreadsPerCore for node %s\n", node)
		}

		reason := ndata["Reason"]

		if brief {
			data = append(data, []string{
				nname,
				node,
				partitions,
				state,
				reason,
			})
		} else {
			data = append(data, []string{
				nname,
				node,
				partitions,
				state,
				version,
				cfgTres,
				allocTres,
				sockets,
				boards,
				tpc,
				reason,
			})
		}
	}

	table := tablewriter.NewWriter(os.Stdout)

	if brief {
		table.SetHeader([]string{"Node", "Hostname", "Partition", "State", "Reason"})
	} else {
		table.SetHeader([]string{"Node", "Hostname", "Partition", "State", "SLURM version", "TRES (configured)", "TRES (allocated)", "Sockets", "Boards", "Threads per core", "Reason"})
	}

	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)

	if !brief {
		table.SetFooter([]string{
			"Sum",
			"",
			fmt.Sprintf("Idle: %d", idleCount),
			fmt.Sprintf("Mixed: %d", mixedCount),
			fmt.Sprintf("Allocated: %d", allocCount),
			fmt.Sprintf("Reserved: %d", reservedCount),
			fmt.Sprintf("Draining: %d", drainingCount),
			fmt.Sprintf("Drained: %d", drainedCount),
			fmt.Sprintf("Down: %d", downCount),
			fmt.Sprintf("Other: %d", otherCount),
			fmt.Sprintf("Total: %d", totalCount),
		})
		table.SetFooterAlignment(tablewriter.ALIGN_LEFT)
	}

	table.AppendBulk(data)
	table.Render()
}

func printReservationStatus(reservation []reservationData, brief bool) {
	var data [][]string
	var nodesCnt uint64
	var coresCnt uint64
	var parts = make(map[string]interface{})
	var activeCnt uint64
	var otherCnt uint64

	for _, rsvData := range reservation {
		rsv, found := rsvData["ReservationName"]
		if !found {
			log.Panicf("BUG: ReservationName not found for reservation data: %+v", rsvData)
		}

		partition, found := rsvData["PartitionName"]
		if !found {
			log.Panicf("BUG: PartitionName not found for reservation %s", rsv)
		}
		parts[partition] = nil

		state, found := rsvData["State"]
		if !found {
			log.Panicf("BUG: State not found for reservation %s", rsv)
		}

		if state == "ACTIVE" {
			activeCnt++
		} else {
			otherCnt++
		}

		startTime, found := rsvData["StartTime"]
		if !found {
			log.Panicf("BUG: StartTime not found for reservation %s", rsv)
		}

		endTime, found := rsvData["EndTime"]
		if !found {
			log.Panicf("BUG: EndTime not found for reservation %s", rsv)
		}

		duration, found := rsvData["Duration"]
		if !found {
			log.Panicf("BUG: Duration not found for reservation %s", rsv)
		}

		nodes, found := rsvData["Nodes"]
		if !found {
			log.Panicf("BUG: Nodes not found for reservation %s", rsv)
		}

		nodeCount, found := rsvData["NodeCnt"]
		if !found {
			log.Panicf("BUG: NodeCnt not found for reservation %s", rsv)
		}
		_nodeCount, err := strconv.ParseUint(nodeCount, 10, 64)
		if err != nil {
			log.Panicf("BUG: Can't convert NodeCnt %s to an integer for reservation %s: %s", nodeCount, rsv, err)
		}
		nodesCnt += _nodeCount

		coreCount, found := rsvData["CoreCnt"]
		if !found {
			log.Panicf("BUG: CoreCnt not found for reservation %s", rsv)
		}
		_coreCount, err := strconv.ParseUint(coreCount, 10, 64)
		if err != nil {
			log.Panicf("BUG: Can't convert CoreCnt %s to an integer for reservation %s: %s", coreCount, rsv, err)
		}
		coresCnt += _coreCount

		features, found := rsvData["Features"]
		if !found {
			log.Panicf("BUG: Features not found for reservation %s", rsv)
		}
		if features == "(null)" {
			features = ""
		}

		flags, found := rsvData["Flags"]
		if !found {
			log.Panicf("BUG: Flags not found for reservation %s", rsv)
		}

		tres, found := rsvData["TRES"]
		if !found {
			log.Panicf("BUG: TRES not fund for reservation %s", rsv)
		}

		users, found := rsvData["Users"]
		if !found {
			log.Panicf("BUG: Users not found for reservation %s", rsv)
		}
		if users == "(null)" {
			users = ""
		}

		accounts, found := rsvData["Accounts"]
		if !found {
			log.Panicf("BUG: Accounts not found for reservation %s", rsv)
		}
		if accounts == "(null)" {
			accounts = ""
		}

		licenses, found := rsvData["Licenses"]
		if !found {
			log.Panicf("BUG: Licenses not found for reservation %s", rsv)
		}
		if licenses == "(null)" {
			licenses = ""
		}

		burstBuffer, found := rsvData["BurstBuffer"]
		if !found {
			log.Panicf("BUG: BurstBuffer not found for reservation %s", rsv)
		}
		if burstBuffer == "(null)" {
			burstBuffer = ""
		}

		watts, found := rsvData["Watts"]
		if !found {
			log.Panicf("BUG: Watts not found for reservation %s", rsv)
		}

		if brief {
			data = append(data, []string{
				rsv,
				partition,
				state,
				startTime,
				endTime,
				duration,
				nodes,
				users,
			})
		} else {
			data = append(data, []string{
				rsv,
				partition,
				state,
				startTime,
				endTime,
				duration,
				nodes,
				nodeCount,
				coreCount,
				features,
				flags,
				tres,
				users,
				accounts,
				licenses,
				burstBuffer,
				watts,
			})
		}
	}

	table := tablewriter.NewWriter(os.Stdout)

	if brief {
		table.SetHeader([]string{"Name", "Partition", "State", "StartTime", "EndTime", "Duration", "Nodes", "Users"})
	} else {
		table.SetHeader([]string{"Name", "Partition", "State", "StartTime", "EndTime", "Duration", "Nodes", "Node count", "Core count", "Features", "Flags", "TRES", "Users", "Accounts", "Licenses", "Burst buffer", "Watts"})
	}

	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)

	if !brief {
		table.SetFooter([]string{
			"Sum",
			fmt.Sprintf("Active: %d", activeCnt),
			fmt.Sprintf("Other: %d", otherCnt),
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			"",
			fmt.Sprintf("Nodes: %d", nodesCnt),
			fmt.Sprintf("Cores: %d", coresCnt),
			fmt.Sprintf("Partitions: %d", len(parts)),
		})
		table.SetFooterAlignment(tablewriter.ALIGN_LEFT)
	}

	table.AppendBulk(data)
	table.Render()
}
